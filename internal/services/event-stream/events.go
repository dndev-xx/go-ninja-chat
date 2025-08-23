package eventstream

import (
	"time"
)

type Event interface {
	eventMarker()
	Validate() error
}

type event struct{}         //
func (*event) eventMarker() {}

// MessageSentEvent indicates that the message was checked by AFC
// and was sent to the manager. Two gray ticks.
type MessageSentEvent struct {
	event
	MessageID   string    `json:"message_id"`
	ChatID      string    `json:"chat_id"`
	UserID      string    `json:"user_id"`
	Content     string    `json:"content"`
	SentAt      time.Time `json:"sent_at"`
	MessageType string    `json:"message_type,omitempty"`
}

func (e MessageSentEvent) Validate() error {
	return nil
}
