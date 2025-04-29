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
	ErrChatNotCreated    = errors.New("chat not created")
	ErrProblemNotCreated = errors.New("problem not created")
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

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	msgRepo 	messagesRepository 		`option:"mandatory" validate:"required"`
	chatRepo 	chatsRepository 		`option:"mandatory" validate:"required"`
	problemRepo problemsRepository		`option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	return UseCase{Options: opts}, opts.Validate()
}

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	// FIXME: 1) Если запрос невалиден, то возвращаем ErrInvalidRequest

	// FIXME: 2) В транзакции:
	// FIXME: 	- если есть сообщение с таким же request_id, то возвращаем его, иначе
	// FIXME:	- создаём чат (если нужно), при ошибке возвращаем ErrChatNotCreated
	// FIXME:	- создаём проблему (если нужно), при ошибке возвращаем ErrProblemNotCreated
	// FIXME:	- создаём новое сообщение

	// FIXME: 3) После коммита транзакции формируем Response.
	return Response{}, nil
}
