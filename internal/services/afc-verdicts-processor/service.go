package afcverdictsprocessor

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/golang-jwt/jwt"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	messagesrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/service_mock.gen.go -package=afcverdictsprocessormocks
type messagesRepository interface {
	MarkAsVisibleForManager(ctx context.Context, msgID types.MessageID) error
	BlockMessage(ctx context.Context, msgID types.MessageID) error
}

type outboxService interface {
	Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
	PublishEvent(ctx context.Context, reqID types.RequestID, msg messagesrepo.Message) error
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	backoffInitialInterval time.Duration `default:"100ms" validate:"min=50ms,max=1s"`
	backoffMaxElapsedTime  time.Duration `default:"5s" validate:"min=500ms,max=1m"`

	brokers         []string `option:"mandatory" validate:"min=1"`
	consumers       int      `option:"mandatory" validate:"min=1,max=16"`
	consumerGroup   string   `option:"mandatory" validate:"required"`
	verdictsTopic   string   `option:"mandatory" validate:"required"`
	verdictsSignKey string

	readerFactory KafkaReaderFactory `option:"mandatory" validate:"required"`
	dlqWriter     KafkaDLQWriter     `option:"mandatory" validate:"required"`

	txtor   transactor         `option:"mandatory" validate:"required"`
	msgRepo messagesRepository `option:"mandatory" validate:"required"`
	outBox  outboxService      `option:"mandatory" validate:"required"`
}

type Service struct {
	Options
	RsaPubKey           *rsa.PublicKey
	r                   KafkaReader
	w                   KafkaDLQWriter
	isNotRetriableError func(err error) bool
}

func New(opts Options) (*Service, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("opts validate error: %w", err)
	}
	service := &Service{
		Options: opts,
		isNotRetriableError: func(err error) bool {
			// Not-retriable ошибки:
			// 1. Ошибки парсинга JSON
			// 2. Ошибки валидации
			// 3. Невалидный JWT

			// Retriable ошибки:
			// 1. Ошибки БД (включая context.Canceled из транзакций)
			// 2. Ошибки сети
			// 3. Любые другие временные ошибки
			if err == nil {
				return false
			}

			errStr := err.Error()

			if strings.Contains(errStr, "invalid JSON") ||
				strings.Contains(errStr, "invalid verdict") ||
				strings.Contains(errStr, "invalid JWT") ||
				strings.Contains(errStr, "backoff timeout exceeded after") ||
				strings.Contains(errStr, "unexpected signing method") {
				return true
			}

			return false
		},
	}
	if strings.TrimSpace(opts.verdictsSignKey) != "" {
		rsaPubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(opts.verdictsSignKey))
		if err != nil {
			return nil, fmt.Errorf("parse RSA public key error: %w", err)
		}
		service.RsaPubKey = rsaPubKey
	}
	service.r = opts.readerFactory(opts.brokers, opts.consumerGroup, opts.verdictsTopic)
	service.w = opts.dlqWriter
	return service, nil
}

func (s *Service) Run(ctx context.Context) error {
	zap.L().Info("Run started")
	defer func() {
		if err := s.r.Close(); err != nil {
			zap.L().Error("failed to close reader", zap.Error(err))
		}
		if err := s.w.Close(); err != nil {
			zap.L().Error("failed to close dlq writer", zap.Error(err))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		m, err := s.r.FetchMessage(ctx)
		if err != nil {
			if err == io.EOF || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}
		if err := s.handleMessage(ctx, m); err != nil {
			zap.L().Error("failed to handle message", zap.Error(err), zap.ByteString("value", m.Value))
			if s.isNotRetriableError(err) {
				if dlqErr := s.sendToDLQ(ctx, m); dlqErr != nil {
					zap.L().Error("failed to send to DLQ", zap.Error(dlqErr))
					return nil
				}
			}
		}

		if err := s.r.CommitMessages(ctx, m); err != nil {
			zap.L().Error("failed to commit message", zap.Error(err))
			return fmt.Errorf("commit message: %w", err)
		}
	}
}

func (s *Service) handleMessage(ctx context.Context, m kafka.Message) error {
	var verdict struct {
		ChatID    string `json:"chatId" validate:"required"`
		MessageID string `json:"messageId" validate:"required"`
		Status    string `json:"status" validate:"required,oneof=ok suspicious"`
	}

	if err := json.Unmarshal(m.Value, &verdict); err != nil {
		if s.RsaPubKey != nil {
			token, jwtErr := jwt.Parse(string(m.Value), func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return s.RsaPubKey, nil
			})

			if jwtErr != nil {
				return fmt.Errorf("invalid JSON or JWT: %w", err)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				verdict.ChatID, _ = claims["chatId"].(string)
				verdict.MessageID, _ = claims["messageId"].(string)
				verdict.Status, _ = claims["status"].(string)
			} else {
				return errors.New("invalid JWT claims")
			}
		} else {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	}

	if err := validator.Validator.Struct(verdict); err != nil {
		return fmt.Errorf("invalid verdict: %w", err)
	}

	msgID := types.MustParse[types.MessageID](verdict.MessageID)
	var operation func(context.Context) error
	if verdict.Status == "ok" {
		operation = func(opCtx context.Context) error {
			return s.txtor.RunInTx(opCtx, func(txCtx context.Context) error {
				if err := s.msgRepo.MarkAsVisibleForManager(txCtx, msgID); err != nil {
					return err
				}
				_, err := s.outBox.Put(txCtx, "client-message-sent", msgID.String(), time.Now())
				return err
			})
		}
	} else if verdict.Status == "suspicious" {
		operation = func(opCtx context.Context) error {
			return s.txtor.RunInTx(opCtx, func(txCtx context.Context) error {
				if err := s.msgRepo.BlockMessage(txCtx, msgID); err != nil {
					return err
				}
				_, err := s.outBox.Put(txCtx, "client-message-blocked", msgID.String(), time.Now())
				return err
			})
		}
	}

	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.InitialInterval = s.backoffInitialInterval
	backoffConfig.MaxElapsedTime = s.backoffMaxElapsedTime

	var lastError error
	attempt := 0
	backoffOps := func() error {
		attempt++
		zap.L().Info("Attempt", zap.Int("attempt", attempt))
		err := operation(ctx)
		lastError = err
		if err == nil {
			return nil
		}
		if s.isNotRetriableError(err) {
			return backoff.Permanent(err)
		}

		zap.L().Info("Retriable error, will retry", zap.Error(err))
		return err
	}
	err := backoff.Retry(backoffOps, backoffConfig)
	if err != nil {
		if perr, ok := err.(*backoff.PermanentError); ok {
			return perr.Unwrap()
		}

		return fmt.Errorf("backoff timeout exceeded after %d attempts: %w", attempt, lastError)
	}
	return nil
}

func (s *Service) sendToDLQ(ctx context.Context, msg kafka.Message) error {
	return s.w.WriteMessages(ctx, msg)
}
