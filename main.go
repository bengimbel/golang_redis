package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/bengimbel/go_redis_api/internal/application"
)

// Main entry point for our application.
// Creating a new app intance, and starting the app
func main() {
	app := application.NewApp()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Failed to start app", err)
	}
}
