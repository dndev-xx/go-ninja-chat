package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"context"

	application "github.com/dndev-xx/go-ninja-chat/pkg/context"
	"go.uber.org/zap"
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
		fmt.Printf("Failed to build app: %v\n", err)
		os.Exit(1)
	}
	
	go func(){ if err := app.DebugServer.Run(ctx); err != nil {
		app.Logger.Error("Error running server", zap.String("server", "debug"), zap.String("err", err.Error()))
	}}()
	go func(){ if err := app.ClientServer.Run(ctx); err != nil {
		app.Logger.Error("Error running server", zap.String("server", "client"), zap.String("err", err.Error()))
	}}()
	
	<-ctx.Done()
	fmt.Println("Shutting down gracefully...")
}

