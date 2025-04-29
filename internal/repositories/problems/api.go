package problems

import (
	"context"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error) {
	// FIXME
	return types.ProblemIDNil, nil
}
