package msgproducer

import (
	"github.com/segmentio/kafka-go"

	"github.com/dndev-xx/go-ninja-chat/internal/logger"
)

const serviceName = "msg-producer"

func NewKafkaWriter(brokers []string, topic string, batchSize int) KafkaWriter {
	if topic == "" {
		topic = serviceName
	}
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     NewChatBalancer(10),
		BatchSize:    batchSize,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Logger:       logger.NewKafkaAdapted(),
		ErrorLogger:  logger.NewKafkaAdapted(),
	}
}
