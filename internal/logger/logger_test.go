package logger_test

import (
	"time"
	"go.uber.org/zap"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"
)

func ExampleInit() {
	if err := logger.Init(logger.NewOptions(
		"error",
		logger.WithProductionMode(false),
		logger.WithClock(fakeClock{}),
	)); err != nil {
		panic(err)
	}

	zap.L().Named("user-cache").Error("inconsistent state", zap.String("uid", "1234"))

	// Output:
	// {"level":"ERROR","T":"2024-01-01T00:00:01.000Z","component":"user-cache","msg":"inconsistent state","uid":"1234"}
}

type fakeClock struct{}

func (f fakeClock) Now() time.Time {
    return time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC)
}

func (f fakeClock) NewTicker(d time.Duration) *time.Ticker {
    return time.NewTicker(d)
}