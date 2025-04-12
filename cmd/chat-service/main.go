package main

import (
	"fmt"
	application "github.com/dndev-xx/go-ninja-chat/pkg/context"
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

	if err := app.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
