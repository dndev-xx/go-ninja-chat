package messages

import (
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type Message struct {
	ID                   types.MessageID
	ChatID               types.ChatID
	ProblemID            types.ProblemID
	AuthorID             types.UserID
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

func adaptStoreMessages(entities []*store.Message) []Message {
	rsl := make([]Message, len(entities))
	for _, msg := range entities {
		rsl = append(rsl, adaptStoreMessage(msg))
	}
	return rsl
}
