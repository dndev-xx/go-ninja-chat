package websocketstream

import (
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

// BaseEvent базовая структура для всех входящих событий
type BaseEvent struct {
	EventType string          `json:"eventType"`
	EventID   types.EventID   `json:"eventId,omitempty"`
	MessageID types.MessageID `json:"messageId,omitempty"`
	RequestID types.RequestID `json:"requestId,omitempty"`
}

// MessageSentEvent входящее событие от клиента (по OpenAPI спецификации)
type IncomingMessageSentEvent struct {
	BaseEvent
	AuthorID  types.UserID `json:"authorId"`
	Body      string       `json:"body"`
	CreatedAt time.Time    `json:"createdAt"`
	IsService bool         `json:"isService"`
}
