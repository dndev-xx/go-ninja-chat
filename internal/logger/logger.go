package logger

import (
	"errors"
	"fmt"
	stdlog "log"
	"syscall"
	"os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate options-gen -out-filename=logger_options.gen.go -from-struct=Options -defaults-from=var
type Options struct {
	level          string `option:"mandatory" validate:"required,oneof=debug info warn error"`
	productionMode bool
	clock          zapcore.Clock
}

type Option func(*Options)

var defaultOptions = Options{
	clock: zapcore.DefaultClock,
}

func NewOptions(level string, opt ...Option) Options {
	options := defaultOptions
	options.level = level

	for _, opt := range opt {
		opt(&options)
	}
	return options
}

func MustInit(opts Options) {
	if err := Init(opts); err != nil {
		panic(err)
	}
}

func Init(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validate options: %v", err)
	}

	level, err := zapcore.ParseLevel(opts.level)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "level",
		NameKey:        "component",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)

	logger := zap.New(
		core,
		zap.WithClock(opts.clock),
	)
	zap.ReplaceGlobals(logger)

	return nil
}

func (o *Options) Validate() error {
	return nil
}

func Sync() {
	if err := zap.L().Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		stdlog.Printf("cannot sync logger: %v", err)
	}
}

func WithProductionMode(productionMode bool) Option {
	return func(o *Options) {
		o.productionMode = productionMode
	}
}

func WithClock(clock zapcore.Clock) Option {
	return func(o *Options) {
		o.clock = clock
	}
}
