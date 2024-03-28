package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bengimbel/go_redis_api/src/httpWeatherClient"
	"github.com/redis/go-redis/v9"
)

const (
	REDIS_ADDR  string = "redis:6379"
	SERVER_ADDR string = ":8080"
)

type App struct {
	Router            http.Handler
	Rdb               *redis.Client
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

// Create a new App instance
func NewApp() *App {
	app := &App{
		Rdb: redis.NewClient(&redis.Options{
			Addr: REDIS_ADDR,
		}),
		WeatherHTTPClient: httpWeatherClient.NewHttpClient(),
	}
	app.LoadApiRoutes()

	return app
}

// Start our App
func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    SERVER_ADDR,
		Handler: a.Router,
	}

	// Ping Redis to make sure we are connected
	err := a.Rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Server failed to connect to redis: %w", err)
	}

	// Gracefully shut down redis.
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

	// If there is an error from channel, select it and
	// handle it gracefully
	select {
	case err = <-channel:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
