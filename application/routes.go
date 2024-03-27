package application

import (
	"github.com/bengimbel/go_redis_api/handler"
	"github.com/bengimbel/go_redis_api/repository"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (a *App) LoadApiRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/api", a.LoadWeatherRouteGroup)

	a.Router = router
}

func (a *App) LoadWeatherRouteGroup(router chi.Router) {
	handler := handler.WeatherHandler{
		Repo: &repository.RedisRepo{
			Client: a.Rdb,
		},
		WeatherHTTPClient: a.WeatherHTTPClient,
	}

	router.Get("/weather", handler.HandleGetWeather)
	router.Get("/weather/cached", handler.HandleGetCachedWeather)
}
