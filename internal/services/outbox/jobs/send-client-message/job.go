package sendclientmessagejob

import (
	"context"
	"fmt"
	"time"

	messagesrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	msgproducer "github.com/dndev-xx/go-ninja-chat/internal/services/msg-producer"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/job_mock.gen.go -package=sendclientmessagejobmocks
const Name = "send-client-message"

type messageProducer interface {
	ProduceMessage(ctx context.Context, message msgproducer.Message) error
}

type messageRepository interface {
	GetMessageByID(ctx context.Context, msgID types.MessageID) (*messagesrepo.Message, error)
}

//go:generate options-gen -out-filename=job_options.gen.go -from-struct=Options
type Options struct {
	messageProducer messageProducer   `option:"mandatory"`
	messageRepo     messageRepository `option:"mandatory"`
}

type Job struct {
	Options
}

func New(opts Options) (*Job, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}
	return &Job{
		Options: opts,
	}, nil
}

func (j *Job) Name() string {
	return Name
}

func (j *Job) Handle(ctx context.Context, payload string) error {
	// FIXME: 1) Заанмаршалили payload
	// FIXME: 2) Достали сообщеньку по MessageID из payload
	// FIXME: 3) Запродьюсили её через messageProducer
	// FIXME: 4) Не забыли приправить логами
	msgID, err := UnmarshalPayload(payload)
	if err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	msg, err := j.messageRepo.GetMessageByID(ctx, msgID)
	if err != nil {
		return fmt.Errorf("get message by id: %w", err)
	}

	err = j.messageProducer.ProduceMessage(ctx, msgproducer.Message{
		ID:         msg.ID,
		ChatID:     msg.ChatID,
		Body:       msg.Body,
		FromClient: true, // FIXME: 5) Поставить правильное значение
	})
	if err != nil {
		return fmt.Errorf("produce message: %w", err)
	}

	return nil
}

func (j *Job) ExecutionTimeout() time.Duration {
	return 30 * time.Second
}

func (j *Job) MaxAttempts() int {
	return 30
}
