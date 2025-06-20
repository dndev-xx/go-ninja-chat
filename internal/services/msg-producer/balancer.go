package msgproducer

import (
	"hash/crc32"

	"github.com/segmentio/kafka-go"
)

// ChatBalancer реализует интерфейс Balancer, обеспечивая:
// - Распределение по партициям на основе chatID
// - Гарантию, что сообщения одного чата попадают в одну партицию
type ChatBalancer struct {
	partitionCount int
}

func NewChatBalancer(partitionCount int) *ChatBalancer {
	return &ChatBalancer{
		partitionCount: partitionCount,
	}
}

func (b *ChatBalancer) Balance(msg kafka.Message, partitions ...int) int {
	if len(partitions) == 0 {
		return 0
	}

	chatID := string(msg.Key)
	if chatID == "" {
		return partitions[0]
	}

	hash := int(crc32.ChecksumIEEE([]byte(chatID)))
	partition := partitions[hash%len(partitions)]
	return partition
}
