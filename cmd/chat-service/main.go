package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	application "github.com/dndev-xx/go-ninja-chat/pkg/context"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := application.NewAppBuilder().
		WithContext(ctx).
		WithConfig().
		WithLogger().
		WithSwagger().
		WithStoresDB().
		WithDebugHTTPSrv().
		WithManagerHTTPSrv().
		WithClientHTTPSrv().
		GetContext()
	if err != nil {
		log.Fatalf("Failed to build app: %v\n", err)
		os.Exit(1)
	}

	defer app.Stores.Close()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return app.DebugServer.Run(ctx) })
	eg.Go(func() error { return app.ClientServer.Run(ctx) })
	eg.Go(func() error { return app.ManagerServer.Run(ctx) })

	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("run app: %v", err)
	}
	zap.L().Info("Shutting down gracefully...")
}
