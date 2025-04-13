package context

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/dndev-xx/go-ninja-chat/internal/config"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	serverdebug "github.com/dndev-xx/go-ninja-chat/internal/server-debug"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

type AppContext struct {
	Config      *config.Config
	Logger      *zap.Logger
	DebugServer *serverdebug.Server
}

type AppBuilder struct {
	App *AppContext
	err error
}

func NewAppBuilder() *AppBuilder {
	return &AppBuilder{
		App: &AppContext{},
	}
}

func (b *AppBuilder) WithConfig() Builder {
	flag.Parse()
	cfg, err := config.ParseAndValidate(*configPath)
	if err != nil {
		b.err = fmt.Errorf("parse and validate config %q: %v", *configPath, err)
		return b
	}
	b.App.Config = cfg
	return b
}

func (b *AppBuilder) WithLogger() Builder {
	if err := logger.Init(logger.NewOptions(
		b.App.Config.Log.Level,
		logger.WithProductionMode(b.App.Config.Global.IsProduction()),
	)); err != nil {
		b.err = err
		return b
	}
	b.App.Logger = logger.GetLogger()
	b.App.Logger.Info("configuration logger was successful with level", zap.String("level", b.App.Config.Log.Level))
	return b
}

func (b *AppBuilder) WithDebugHTTPSrv() Builder {
	srvDebug, err := serverdebug.New(serverdebug.NewOptions(b.App.Config.Servers.Debug.Addr))
	if err != nil {
		b.err = fmt.Errorf("init debug server: %v", err)
		return b
	}
	b.App.DebugServer = srvDebug
	b.App.Logger.Info("configuration debug server was successful")
	return b
}

func (b *AppBuilder) GetContext() (*AppContext, error) {
	return b.App, b.err
}

func (a *AppContext) DebugServerRun() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return a.DebugServer.Run(ctx)
	})
	return eg.Wait()
}
