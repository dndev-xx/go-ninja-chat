package problems

import (
	"context"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error) {
	id, err := r.db.Problem(ctx).
		Create().
		SetChatID(chatID).
		OnConflict().
		UpdateNewValues().
		ID(ctx)
	if err != nil {
		return types.ProblemIDNil, err
	}
	return id, nil
}
