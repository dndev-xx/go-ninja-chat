package messages

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) MarkAsVisibleForManager(ctx context.Context, msgID types.MessageID) error {
	query := `
		update messages set is_visible_for_manager = true 
		where id = $1 
		returning id;
	`
	id, err := r.executeQuery(ctx, msgID, query)
	if err != nil {
		return err
	}
	zap.L().Info("success mark visible manager with id:", zap.String("id", id.String()))
	return nil
}

func (r *Repo) BlockMessage(ctx context.Context, msgID types.MessageID) error {
	query := `
		update messages set is_blocked = true 
		where id = $1 
		returning id;
	`
	id, err := r.executeQuery(ctx, msgID, query)
	if err != nil {
		return err
	}
	zap.L().Info("success mark blocked msg with id:", zap.String("id", id.String()))
	return nil
}

func (r *Repo) executeQuery(ctx context.Context, msgID types.MessageID, query string) (*types.MessageID, error) {
	rows, err := r.db.Message(ctx).QueryContext(ctx, query, msgID)
	if err != nil {
		return nil, fmt.Errorf("query context: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows err: %w", err)
		}
	}
	var id types.MessageID
	if err := rows.Scan(&id); err != nil {
		return nil, fmt.Errorf("scan messages: %w", err)
	}
	return &id, nil
}
