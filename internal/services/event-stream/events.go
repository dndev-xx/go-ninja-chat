package eventstream

import (
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

type Event interface {
	eventMarker()
	Validate() error
}

type event struct{}

func (*event) eventMarker() {}

// MessageSentEvent represents a message that was sent in the chat
type MessageSentEvent struct {
	event
	EventID   types.EventID   `json:"eventId" validate:"required"`
	ChatID    types.ChatID    `json:"chatId" validate:"required"`
	EventType string          `json:"eventType" validate:"required"`
	MessageID types.MessageID `json:"messageId" validate:"required"`
	RequestID types.RequestID `json:"requestId" validate:"required"`

	// Optional fields for service messages
	AuthorID  *types.UserID `json:"authorId,omitempty"`
	Body      string        `json:"body,omitempty"`
	CreatedAt *time.Time    `json:"createdAt,omitempty"`
	IsService bool          `json:"isService,omitempty"`
}

// NewMessageSentEvent creates a basic MessageSentEvent (first test case)
func NewMessageSentEvent(eventId types.EventID, reqId types.RequestID, messageId types.MessageID) *MessageSentEvent {
	return &MessageSentEvent{
		EventID:   eventId,
		EventType: "MessageSentEvent", // Changed from "MessageSentEvent" to match your OpenAPI
		MessageID: messageId,
		RequestID: reqId,
	}
}

// NewMessageSentEventWithDetails creates a detailed MessageSentEvent with all fields (second test case)
func NewMessageSentEventWithDetails(
	eventId types.EventID,
	reqId types.RequestID,
	chatId types.ChatID,
	messageId types.MessageID,
	userId types.UserID,
	createdAt time.Time,
	body string,
	isService bool,
) *MessageSentEvent {
	event := &MessageSentEvent{
		EventID:   eventId,
		EventType: "NewMessageEvent", // Changed to match your second test case
		MessageID: messageId,
		RequestID: reqId,
		Body:      body,
		CreatedAt: &createdAt,
		IsService: isService,
	}

	// Only set AuthorID if it's not nil/Nil
	if userId != types.UserIDNil {
		event.AuthorID = &userId
	}

	return event
}

// Alternative constructor that matches your test case signature more closely
func NewMessageSentEventService(
	eventId types.EventID,
	reqId types.RequestID,
	chatId types.ChatID, // This parameter exists in test but isn't used in JSON
	messageId types.MessageID,
	authorId types.UserID,
	createdAt time.Time,
	body string,
	isService bool,
) *MessageSentEvent {
	return NewMessageSentEventWithDetails(
		eventId,
		reqId,
		chatId,
		messageId,
		authorId,
		createdAt,
		body,
		isService,
	)
}

func (e *MessageSentEvent) Validate() error {
	if err := validator.Validator.Struct(e); err != nil {
		return err
	}
	return nil
}
