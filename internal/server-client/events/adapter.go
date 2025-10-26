package clientevents

import (
	"time"

	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	websocketstream "github.com/dndev-xx/go-ninja-chat/internal/websocket-stream"
)

var _ websocketstream.EventAdapter = Adapter{}

type Adapter struct{}

type MessageSentEventDTO struct {
	EventID   types.EventID   `json:"eventId"`
	EventType string          `json:"eventType"`
	MessageID types.MessageID `json:"messageId"`
	RequestID types.RequestID `json:"requestId"`
}

type NewMessageEventDTO struct {
	EventID   types.EventID   `json:"eventId"`
	EventType string          `json:"eventType"`
	MessageID types.MessageID `json:"messageId"`
	RequestID types.RequestID `json:"requestId"`
	Body      string          `json:"body,omitempty"`
	AuthorID  types.UserID    `json:"authorId,omitempty"`
	CreatedAt time.Time       `json:"createdAt,omitempty"`
	IsService bool            `json:"isService,omitempty"`
}

func (Adapter) Adapt(ev eventstream.Event) (any, error) {
	msgEvent, ok := ev.(*eventstream.MessageSentEvent)
	if !ok {
		return nil, nil
	}

	switch msgEvent.EventType {
	case "MessageSentEvent":
		return MessageSentEventDTO{
			EventID:   msgEvent.EventID,
			EventType: msgEvent.EventType,
			MessageID: msgEvent.MessageID,
			RequestID: msgEvent.RequestID,
		}, nil

	case "NewMessageEvent":
		dto := NewMessageEventDTO{
			EventID:   msgEvent.EventID,
			EventType: msgEvent.EventType,
			MessageID: msgEvent.MessageID,
			RequestID: msgEvent.RequestID,
			AuthorID:  *msgEvent.AuthorID,
			Body:      msgEvent.Body,
			IsService: msgEvent.IsService,
		}

		if msgEvent.CreatedAt != nil {
			dto.CreatedAt = *msgEvent.CreatedAt
		}
		return dto, nil

	default:
		if msgEvent.Body != "" || msgEvent.CreatedAt != nil || msgEvent.IsService {
			dto := NewMessageEventDTO{
				EventID:   msgEvent.EventID,
				EventType: msgEvent.EventType,
				MessageID: msgEvent.MessageID,
				RequestID: msgEvent.RequestID,
				Body:      msgEvent.Body,
				IsService: msgEvent.IsService,
			}

			if msgEvent.CreatedAt != nil {
				dto.CreatedAt = *msgEvent.CreatedAt
			}

			return dto, nil
		} else {
			return MessageSentEventDTO{
				EventID:   msgEvent.EventID,
				EventType: msgEvent.EventType,
				MessageID: msgEvent.MessageID,
				RequestID: msgEvent.RequestID,
			}, nil
		}
	}
}
