package messages

import (
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type Message struct {
	ID                  types.MessageID `json:"id"`
	ChatID              types.ChatID    `json:"chatId"`
	ProblemID           types.ProblemID `json:"problemId"`
	AuthorID            types.UserID    `json:"authorId"`
	Body                string          `json:"body"`
	IsVisibleForClient  bool            `json:"isVisibleForClient"`
	IsVisibleForManager bool            `json:"isVisibleForManager"`
	IsBlocked           bool            `json:"isBlocked"`
	IsService           bool            `json:"isService"`
	CreatedAt           time.Time       `json:"createdAt"`
	CheckedAt           *time.Time      `json:"checkedAt,omitempty"`
}

func adaptStoreMessage(m *store.Message) Message {
	return Message{
		ID:                  m.ID,
		ChatID:              m.ChatID,
		ProblemID:           m.ProblemID,
		AuthorID:            m.AuthorID,
		Body:                m.Body,
		IsVisibleForClient:  m.IsVisibleForClient,
		IsVisibleForManager: m.IsVisibleForManager,
		IsBlocked:           m.IsBlocked,
		IsService:           m.IsService,
		CreatedAt:           m.CreatedAt,
		CheckedAt:           m.CheckedAt,
	}
}

func adaptStoreMessages(entities []*store.Message) []Message {
	var result []Message
	for _, msg := range entities {
		adapted := adaptStoreMessage(msg)
		result = append(result, adapted)
	}
	return result
}
