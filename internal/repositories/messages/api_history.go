package messages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/store/message"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const (
	minSize = 10
	maxSize = 100
)

var (
	ErrInvalidPageSize = errors.New("invalid page size")
	ErrInvalidCursor   = errors.New("invalid cursor")
)

type Cursor struct {
	LastCreatedAt time.Time
	PageSize      int
}

func (r *Repo) GetClientChatMessages(
	ctx context.Context,
	clientID types.UserID,
	pageSize int,
	cursor *Cursor,
) ([]Message, *Cursor, error) {
	pCurr := getCurrentSize(pageSize)
	var lastCreatedAt time.Time
	if cursor != nil {
		pCurr = getCurrentSize(cursor.PageSize)
		if cursor.LastCreatedAt.IsZero() {
			return nil, nil, ErrInvalidCursor
		}
		lastCreatedAt = cursor.LastCreatedAt
	}
	query := r.db.Message(ctx).
		Query().
		Where(
			message.AuthorIDEQ(clientID),
			message.IsVisibleForClient(true),
		).
		Order(store.Desc(message.FieldCreatedAt)).
		Limit(pCurr + 1)

	if !lastCreatedAt.IsZero() {
		query = query.Where(message.CreatedAtLT(lastCreatedAt))
	}
	messages, err := query.All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	var nextCursor *Cursor
	if len(messages) > pCurr {
		lastMsg := messages[pCurr-1]
		nextCursor = &Cursor{
			LastCreatedAt: lastMsg.CreatedAt,
			PageSize:      pCurr,
		}
		messages = messages[:pCurr]
	}

	return adaptStoreMessages(messages), nextCursor, nil
}

func getCurrentSize(size int) int {
	pCurr := minSize
	if size >= minSize && size <= maxSize {
		pCurr = size
	}
	return pCurr
}
