package logger

import (
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	_        kafka.Logger = (*KafkaAdapted)(nil)
	instance *KafkaAdapted
	once     sync.Once
)

type KafkaAdapted struct {
	logger *zap.Logger
}

func NewKafkaAdapted() *KafkaAdapted {
	once.Do(func() {
		instance = &KafkaAdapted{
			logger: zap.L(),
		}
	})
	return instance
}

func (k *KafkaAdapted) Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	k.logger.Info(message)
}

func (k *KafkaAdapted) Close() error {
	return k.logger.Sync()
}
