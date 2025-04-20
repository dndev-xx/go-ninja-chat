package messages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/store/message"
	"github.com/google/uuid"
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
	clientID uuid.UUID,
	pageSize int,
	cursor *Cursor,
) ([]Message, *Cursor, error) {
	pCurr := getCurrentSize(pageSize)
	var lastCreatedAt time.Time
	if cursor != nil {
		pCurr = getCurrentSize(cursor.PageSize)
		if cursor.LastCreatedAt.IsZero() {
			return nil, nil, fmt.Errorf("invalid date at cursor")
		}
		lastCreatedAt = cursor.LastCreatedAt
	}
	query := r.db.Message(ctx).
		Query().
		Where(
			message.ChatIDEQ(clientID),
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

	return r.toDto(messages), nextCursor, nil
}

func (r *Repo) toDto(entities []*store.Message) []Message {
	rsl := make([]Message, len(entities))
	for _, msg := range entities {
		rsl = append(rsl, adaptStoreMessage(msg))
	}
	return rsl
}

func getCurrentSize(size int) int {
	pCurr := minSize
	if size >= minSize && size <= maxSize {
		pCurr = size
	}
	return pCurr
}
