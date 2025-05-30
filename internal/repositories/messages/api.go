package messages

import (
	"context"
	"errors"
	"fmt"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/store/message"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var ErrMsgNotFound = errors.New("message not found")
var ErrMsgUpsert = errors.New("message don`t create or update")

func (r *Repo) GetMessageByRequestID(ctx context.Context, reqID types.RequestID) (*Message, error) {
	m, err := r.db.Message(ctx).Query().
		Unique(false).
		Where(message.InitialRequestID(reqID)).
		Only(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return nil, fmt.Errorf("request id %v: %w", reqID, ErrMsgNotFound)
		}
		return nil, fmt.Errorf("query message by request id %v: %v", reqID, err)
	}

	mm := adaptStoreMessage(m)
	return &mm, nil
}

// CreateClientVisible creates a message that is visible only to the client.
func (r *Repo) CreateClientVisible(
	ctx context.Context,
	reqID types.RequestID,
	problemID types.ProblemID,
	chatID types.ChatID,
	authorID types.UserID,
	msgBody string,
) (*Message, error) {
	m, err := r.db.Message(ctx).Create().
		SetChatID(chatID).
		SetProblemID(problemID).
		SetAuthorID(authorID).
		SetIsVisibleForClient(true).
		SetIsVisibleForManager(false).
		SetBody(msgBody).
		SetInitialRequestID(reqID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create msg: %v", err)
	}

	mm := adaptStoreMessage(m)
	return &mm, nil
}
