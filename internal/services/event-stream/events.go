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

type event struct{}         //
func (*event) eventMarker() {}

// MessageSentEvent indicates that the message was checked by AFC
// and was sent to the manager. Two gray ticks.
type MessageSentEvent struct {
	event
	MessageID    types.MessageID `json:"message_id" validate:"required"`  // Уникальный идентификатор сообщения
	ChatID       types.ChatID    `json:"chat_id" validate:"required"`     // Идентификатор чата
	UserID       types.UserID    `json:"user_id" validate:"required"`     // Идентификатор пользователя, отправившего сообщение
	Content      string          `json:"content" validate:"required"`      // Содержимое сообщения
	SentAt       time.Time       `json:"sent_at" validate:"required"`      // Время отправки сообщения
	MessageType  string          `json:"message_type" validate:"max=100,required"` // Тип сообщения (например, текст, изображение и т.д.)
	EventID      types.EventID   `json:"event_id" validate:"required"`     // Идентификатор события
	RequestID    types.RequestID `json:"request_id" validate:"required"`   // Идентификатор запроса
	IsCheckAtAFC bool            `json:"is_check_at_afc"`                  // Флаг проверки на AFC (может быть опциональным)
}

func (e MessageSentEvent) Validate() error {
	if err := validator.Validator.Struct(e); err != nil {
		return err
	}
	return nil
}
