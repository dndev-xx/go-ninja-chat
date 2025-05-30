package sendmessage

import (
	"context"
	"errors"
	"fmt"

	messagesrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=sendmessagemocks
var (
	ErrInvalidRequest    = errors.New("invalid request")
	ErrIdempotencyKey    = errors.New("invalid idempotency key")
	ErrChatNotCreated    = errors.New("chat not created")
	ErrProblemNotCreated = errors.New("problem not created")
	ErrMessageNotCreated = errors.New("message not created")
)

type chatsRepository interface {
	CreateIfNotExists(ctx context.Context, userID types.UserID) (types.ChatID, error)
}

type messagesRepository interface {
	GetMessageByRequestID(ctx context.Context, reqID types.RequestID) (*messagesrepo.Message, error)
	CreateClientVisible(
		ctx context.Context,
		reqID types.RequestID,
		problemID types.ProblemID,
		chatID types.ChatID,
		authorID types.UserID,
		msgBody string,
	) (*messagesrepo.Message, error)
}

type problemsRepository interface {
	CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error)
}

type requestRepository interface {
	CreateIfNotExists(ctx context.Context, requestID types.RequestID) (bool, error)
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	msgRepo 	messagesRepository 		`option:"mandatory" validate:"required"`
	chatRepo 	chatsRepository 		`option:"mandatory" validate:"required"`
	problemRepo problemsRepository		`option:"mandatory" validate:"required"`
	tx			transactor				`option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	return UseCase{Options: opts}, opts.Validate()
}

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	if err := req.Validate(); err != nil {
		return Response{}, fmt.Errorf("validate request: %w: %v", ErrInvalidRequest, err)
	}

	var msg *messagesrepo.Message

	if err := u.tx.RunInTx(ctx, func(ctx context.Context) error {
		m, err := u.msgRepo.GetMessageByRequestID(ctx, req.ID)
		if nil == err {
			msg = m
			return nil
		}
		if !errors.Is(err, messagesrepo.ErrMsgNotFound) {
			return fmt.Errorf("get msg by initial request id: %v", err)
		}

		chatID, err := u.chatRepo.CreateIfNotExists(ctx, req.ClientID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrChatNotCreated, err)
		}

		problemID, err := u.problemRepo.CreateIfNotExists(ctx, chatID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrProblemNotCreated, err)
		}

		m, err = u.msgRepo.CreateClientVisible(ctx, req.ID, problemID, chatID, req.ClientID, req.MessageBody)
		if err != nil {
			return fmt.Errorf("create client visible message: %v", err)
		}

		msg = m
		return nil
	}); err != nil {
		return Response{}, fmt.Errorf("`send client message` tx: %w", err)
	}

	return Response{
		AuthorID:  msg.AuthorID,
		MessageID: msg.ID,
		CreatedAt: msg.CreatedAt,
	}, nil
}
