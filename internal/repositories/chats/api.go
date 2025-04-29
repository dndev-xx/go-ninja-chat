package chatsrepo

import (
	"context"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, userID types.UserID) (types.ChatID, error) {
	return types.ChatIDNil, nil
}
