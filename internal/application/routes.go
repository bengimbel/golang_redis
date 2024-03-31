package application

import (
	"github.com/bengimbel/go_redis_api/internal/handler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// Load routes and bind them to our App struct.
func (a *App) LoadApiRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/api", a.LoadWeatherRouteGroup)

	a.Router = router
}

func (a *App) LoadWeatherRouteGroup(router chi.Router) {
	handler := handler.NewWeatherHandler(a.Rdb)

	router.Get("/weather", handler.HandleRetrieveWeather)
	router.Get("/weather/cached", handler.HandleRetrieveCachedWeather)
}
