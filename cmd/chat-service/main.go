package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"syscall"
	"github.com/dndev-xx/go-ninja-chat/internal/config"
	serverdebug "github.com/dndev-xx/go-ninja-chat/internal/server-debug"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

func main() {
	if err := run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}

func run() (errReturned error) {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseAndValidate(*configPath)
	if err != nil {
		return fmt.Errorf("parse and validate config %q: %v", *configPath, err)
	}

	if err := logger.Init(logger.NewOptions(
		cfg.Log.Level,
		logger.WithProductionMode(false),
	)); err != nil {
		panic(err)
	}

	logger ,err := logger.NewSentryClient(cfg.Sentry.DSN, cfg.Global.Env, "0.0.1")
	if err != nil {
		return fmt.Errorf("init sentry client: %v", err)
	}
	srvDebug, err := serverdebug.New(logger, serverdebug.NewOptions(cfg.Servers.Debug.Addr))
	if err != nil {
		return fmt.Errorf("init debug server: %v", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	// Запуск серверов
	eg.Go(func() error { return srvDebug.Run(ctx) })

	// Запуск сервисов
	// Ждут своего часа.
	// ...

	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("wait app stop: %v", err)
	}

	return nil
}