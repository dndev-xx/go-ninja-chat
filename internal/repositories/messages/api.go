package messages

import (
	"context"
	"errors"

	"github.com/dndev-xx/go-ninja-chat/internal/store/message"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var ErrMsgNotFound = errors.New("message not found")
var ErrMsgUpsert = errors.New("message don`t create or update")

func (r *Repo) GetMessageByRequestID(ctx context.Context, reqID types.RequestID) (*Message, error) {
	return nil, nil
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
	id, err := r.db.Message(ctx).
		Create().
		SetChatID(chatID).
		SetAuthorID(authorID).
		SetBody(msgBody).
		SetProblemID(problemID).
		SetIsVisibleForClient(true).
		OnConflict().
		UpdateNewValues().
		ID(ctx)

	if err != nil {
		return nil, ErrMsgUpsert
	}
	msg, err := r.db.Message(ctx).Query().
		Where(message.ID(id)).First(ctx)
	if err != nil {
		return nil, err
	}
	rsl := adaptStoreMessage(msg)
	return &rsl, nil
}
