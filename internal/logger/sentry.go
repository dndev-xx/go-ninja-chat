package logger

import (
	"os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewSentryClient(dsn, env, version string) (*zap.Logger, error) {
	// Run with sentry only linux machine
	/*err := sentry.Init(sentry.ClientOptions{
        Dsn:         dsn,
        Release:     version,
        Environment: env,
        HTTPTransport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true, //nolint:gosec // non-prod solution
            },
        },
    })
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry client: %w", err)
	}*/

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

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
	)

	logger := zap.New(core)
	logger.Info("Client initialized", zap.String("env", env), zap.String("version", version))

	return logger, nil
}