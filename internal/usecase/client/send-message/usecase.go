package sendmessage

import (
	"context"
	"errors"

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
	requestRepo requestRepository		`option:"mandatory" validate:"required"`
	tx			transactor				`option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	return UseCase{Options: opts}, opts.Validate()
}

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	rsl := Response{}
	if err := req.Validate(); err != nil {
		return rsl, ErrInvalidRequest
	}
	err := u.tx.RunInTx(ctx, func(context.Context) error {
		exists, err := u.requestRepo.CreateIfNotExists(ctx, req.ID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrIdempotencyKey
		}
		chatID, err := u.chatRepo.CreateIfNotExists(ctx, req.ClientID)
		if err != nil {
			return ErrChatNotCreated
		}
		problemID, err := u.problemRepo.CreateIfNotExists(ctx, chatID)
		if err != nil {
			return ErrProblemNotCreated
		}
		msg, err := u.msgRepo.CreateClientVisible(ctx, req.ID, problemID, chatID, req.ClientID, req.MessageBody)
		if err != nil {
			return ErrMessageNotCreated
		}
		rsl.AuthorID = msg.AuthorID
		rsl.MessageID = msg.ID
		rsl.CreatedAt = msg.CreatedAt
		return nil
	})
	if err != nil {
		return rsl, err
	}
	return rsl, nil
}
