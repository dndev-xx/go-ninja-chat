package problems

import (
	"context"
	"fmt"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/store/chat"
	"github.com/dndev-xx/go-ninja-chat/internal/store/problem"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, chatID types.ChatID) (types.ProblemID, error) {
	pID, err := r.db.Problem(ctx).Query().
		Unique(false).
		Where(
			problem.HasChatWith(chat.ID(chatID)),
			problem.ResolvedAtIsNil(),
		).
		FirstID(ctx)
	if nil == err {
		return pID, nil
	}
	if !store.IsNotFound(err) {
		return types.ProblemIDNil, fmt.Errorf("select existent problem: %v", err)
	}

	p, err := r.db.Problem(ctx).Create().
		SetChatID(chatID).
		Save(ctx)
	if err != nil {
		return types.ProblemIDNil, fmt.Errorf("create new problem: %v", err)
	}

	return p.ID, nil
}

func (r *Repo) GetManagerOpenProblemsCount(ctx context.Context, managerID types.UserID) (int, error) {
	count, err := r.db.Problem(ctx).Query().
		Unique(false).
		Where(
			problem.ManagerID(managerID),
			problem.ResolvedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count open problems: %v", err)
	}

	return count, nil
}
