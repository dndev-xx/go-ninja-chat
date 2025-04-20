package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	application "github.com/dndev-xx/go-ninja-chat/pkg/context"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := application.NewAppBuilder().
		WithConfig().
		WithLogger().
		WithDebugHTTPSrv().
		WithSwagger().
		WithClientHTTPSrv().
		GetContext()

	if err != nil {
		log.Fatalf("Failed to build app: %v\n", err)
		os.Exit(1)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return app.DebugServer.Run(ctx) })
	eg.Go(func() error { return app.ClientServer.Run(ctx) })

	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("run app: %v", err)
	}
	log.Println("Shutting down gracefully...")
}
