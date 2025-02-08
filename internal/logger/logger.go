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

	var atomicLevel zap.AtomicLevel
	switch opts.level {
		case "debug":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
        case "info":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
        case "warn":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.WarnLevel)
        case "error":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
        case "dpanic":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.DPanicLevel)
        case "panic":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.PanicLevel)
        case "fatal":
            atomicLevel = zap.NewAtomicLevelAt(zapcore.FatalLevel)
        default:
            return fmt.Errorf("unknown log level: %s", opts.level)
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

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	logger := zap.New(core, zap.WithClock(opts.clock))
	zap.ReplaceGlobals(logger)
	Sync()
	return nil
}

func (o *Options) Validate() error {
	if o.level == "" {
        return errors.New("level is required")
    }
	if _, err := zapcore.ParseLevel(o.level); err!= nil {
        return fmt.Errorf("invalid log level: %w", err)
    }
	if o.clock == nil {
        return errors.New("clock is required")
    }
    if o.productionMode && os.Getenv("GIN_MODE") != "release" {
        return errors.New("production mode is only available in release mode")
    }
	return nil
}

func Sync() {
	if err := zap.L().Sync(); err != nil && !errors.Is(err, syscall.EINVAL) {
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
