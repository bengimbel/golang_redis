package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/bengimbel/go_redis_api/application"
)

func main() {
	app := application.NewApp()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Failed to start app", err)
	}
}
