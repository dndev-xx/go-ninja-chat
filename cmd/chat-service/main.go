package main

import (
	"fmt"
	application "github.com/dndev-xx/go-ninja-chat/pkg/context"
	"go.uber.org/zap"
)

func main() {
	app, err := application.NewAppBuilder().
		WithConfig().
		WithLogger().
		WithDebugHTTPSrv().
		GetContext()

	if err != nil {
		fmt.Printf("Failed to build app: %v\n", err)
		return
	}

	if err := app.DebugServerRun(); err != nil {
		app.Logger.Error("Error running app", zap.String("err", err.Error()))
	}
}
