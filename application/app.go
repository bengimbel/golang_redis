package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bengimbel/go_redis_api/httpWeatherClient"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Router            http.Handler
	Rdb               *redis.Client
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

func NewApp() *App {
	app := &App{
		Rdb: redis.NewClient(&redis.Options{
			Addr: "redis:6379",
		}),
		WeatherHTTPClient: httpWeatherClient.NewHttpClient(),
	}
	app.LoadApiRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.Router,
	}

	err := a.Rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Server failed to connect to redis: %w", err)
	}

	defer func() {
		if err := a.Rdb.Close(); err != nil {
			fmt.Println("Failed to close redis", err)
		}
	}()

	fmt.Println("Starting Server on port 8080")

	// Using buffered channel, only 1 error can happen here
	channel := make(chan error, 1)

	// Go routine to start server on another thread in case for gracefun shutdown
	// Wont block main thread if fails
	go func() {
		if err := server.ListenAndServe(); err != nil {
			channel <- fmt.Errorf("Server failed to start: %w", err)
		}
		close(channel)
	}()

	select {
	case err = <-channel:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
