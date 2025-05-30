package gethistory

import (
	"context"
	"errors"

	"github.com/dndev-xx/go-ninja-chat/internal/cursor"
	messagesrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mock.gen.go -package=gethistorymocks
var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidCursor  = errors.New("invalid cursor")
)

type messagesRepository interface {
	GetClientChatMessages(
		ctx context.Context,
		clientID types.UserID,
		pageSize int,
		cursor *messagesrepo.Cursor,
	) ([]messagesrepo.Message, *messagesrepo.Cursor, error)
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	msgRepo messagesRepository `option:"mandatory" validate:"required"`
}

type UseCase struct {
	Options
}

func New(opts Options) (UseCase, error) {
	if err := opts.Validate(); err != nil {
		return UseCase{}, err
	}
	return UseCase{Options: opts}, nil
}

func (u UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	rsl := Response{}
	var cursorData *messagesrepo.Cursor
	if req.Cursor != "" {
		cursorData = &messagesrepo.Cursor{}
	}
	if err := req.Validate(); err != nil {
		return rsl, ErrInvalidRequest
	}
	if err := cursor.Decode(req.Cursor, cursorData); err != nil {
		return rsl, ErrInvalidCursor
	}
	msgs, curs, err := u.msgRepo.GetClientChatMessages(ctx, req.ClientID, req.PageSize, cursorData)
	if err != nil {
		if errors.Is(err, messagesrepo.ErrInvalidCursor) {
			return rsl, messagesrepo.ErrInvalidCursor
		}
		return rsl, err
	}
	var nextCursor string
	messages := make([]Message, 0)
	if curs != nil {
	   	nextCursor, err = cursor.Encode(curs)
		if err != nil {
			return rsl, err
		}
	}
	rsl.NextCursor = nextCursor
	for _, msg := range msgs {
		isReceived := false
		if msg.IsVisibleForManager && !msg.IsBlocked {
			isReceived = true
		}
		curMsg := Message{
			ID: msg.ID,
			AuthorID: msg.AuthorID,
			Body: msg.Body,
			IsBlocked: msg.IsBlocked,
			IsService: msg.IsService,
			CreatedAt: msg.CreatedAt,
			IsReceived: isReceived,
		}
		messages = append(messages, curMsg)
	}
	rsl.Messages = messages
	return rsl, nil
}
