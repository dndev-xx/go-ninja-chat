package msgproducer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const (
	FALSE = iota
	TRUE
)

type Message struct {
	ID         types.MessageID `json:"id"`
	ChatID     types.ChatID    `json:"chatId"`
	Body       string          `json:"body"`
	FromClient bool            `json:"fromClient"`
}

func (s *Service) ProduceMessage(ctx context.Context, msg Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if s.cipher != nil {
		nonce, err := s.nonceFactory(s.cipher.NonceSize())
		if err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}

		encrypted := s.cipher.Seal(nil, nonce, jsonData, nil)
		jsonData = append(nonce, encrypted...)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(msg.ChatID.String()),
		Value: jsonData,
		Headers: []kafka.Header{
			{
				Key:   "message-id",
				Value: []byte(msg.ID.String()),
			},
			{
				Key:   "from-client",
				Value: []byte{boolToByte(msg.FromClient)},
			},
		},
	}
	if err := s.wr.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func (s *Service) Close() error {
	if err := s.wr.Close(); err != nil {
		return err
	}
	return nil
}

func boolToByte(b bool) byte {
	if b {
		return TRUE
	}
	return FALSE
}
