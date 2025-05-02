package chatsrepo

import (
	"context"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, userID types.UserID) (types.ChatID, error) {
	id, err := r.db.Chat(ctx).
		Create().
		SetClientID(userID).
		OnConflict().
		UpdateNewValues().
		ID(ctx)
	if err != nil {
		return types.ChatIDNil, err
	}
	return id, nil
}
