package application

import (
	"time"

	"github.com/bengimbel/go_redis_api/src/handler"
	"github.com/bengimbel/go_redis_api/src/repository"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/cache/v9"
)

// Load routes and bind them to our App struct.
func (a *App) LoadApiRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/api", a.LoadWeatherRouteGroup)

	a.Router = router
}

// Setting Cache to use local in-process storage
// to cache the small subset of recent keys.
// Key/Values use LRU (least recently used)
// for 1 minute in local in-process storage
// before looking into the Redis Cache.
func (a *App) LoadWeatherRouteGroup(router chi.Router) {
	handler := handler.WeatherHandler{
		Repo: &repository.RedisRepo{
			Cache: cache.New(&cache.Options{
				Redis:      a.Rdb,
				LocalCache: cache.NewTinyLFU(1000, time.Minute),
			}),
		},
		WeatherHTTPClient: a.WeatherHTTPClient,
	}

	router.Get("/weather", handler.HandleRetrieveWeather)
	router.Get("/weather/cached", handler.HandleRetrieveCachedWeather)
}
