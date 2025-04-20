package messages

import (
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/google/uuid"
)

type Message struct {
	ID                   uuid.UUID
	ChatID               uuid.UUID
	ProblemID            uuid.UUID
	AuthorID             uuid.UUID
	Body                 string
	IsVisibleForClient   bool
	IsVisibleForManager  bool
	IsBlocked            bool
	IsService            bool
	CreatedAt            time.Time
	CheckedAt            *time.Time
}

func adaptStoreMessage(m *store.Message) Message {
	return Message{
		ID:                   m.ID,
		ChatID:               m.ChatID,
		ProblemID:            m.ProblemID,
		AuthorID:             m.AuthorID,
		Body:                 m.Body,
		IsVisibleForClient:   m.IsVisibleForClient,
		IsVisibleForManager:  m.IsVisibleForManager,
		IsBlocked:            m.IsBlocked,
		IsService:            m.IsService,
		CreatedAt:            m.CreatedAt,
		CheckedAt:            m.CheckedAt,
	}
}
