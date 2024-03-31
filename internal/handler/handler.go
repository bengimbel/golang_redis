package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/bengimbel/go_redis_api/internal/model"
	"github.com/bengimbel/go_redis_api/internal/service"
	"github.com/bengimbel/go_redis_api/pkg/errorPkg"
	"github.com/redis/go-redis/v9"
)

type WeatherHandler struct {
	Service service.WeatherServiceImplementor
}

func NewWeatherHandler(rds *redis.Client) *WeatherHandler {
	return &WeatherHandler{
		Service: service.NewWeatherService(rds),
	}
}

// Handler for fetching weather from open weather map API.
func (wh *WeatherHandler) HandleRetrieveWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))
	ctx := r.Context()
	var result model.WeatherResponse

	// Check the cache before fetching
	keyExists := wh.Service.DoesKeyExist(ctx, city)
	if keyExists {
		value, err := wh.Service.RetrieveWeatherFromCache(ctx, city)
		if err != nil {
			errorPkg.RenderInternalServerError(w, err)
			return
		}
		result = value
	} else {
		value, err := wh.Service.RetrieveAndCacheWeatherAsync(ctx, city)
		if err != nil {
			errorPkg.RenderBadRequestError(w, err)
			return
		}
		result = value
	}
	// Marshal struct to json for the return.
	// If error while decoding to json,
	// render a general server error
	response, err := json.Marshal(&result)
	if err != nil {
		log.Println("Error decoding response to json", err)
		errorPkg.RenderInternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (wh *WeatherHandler) HandleRetrieveCachedWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))
	ctx := r.Context()

	// Get results from redis cache
	result, err := wh.Service.RetrieveWeatherFromCache(ctx, city)
	if err != nil {
		errorPkg.RenderBadRequestError(w, err)
		return
	}

	// Marshal struct to json for the return.
	// If error while decoding to json,
	// render a general server error
	response, err := json.Marshal(result)
	if err != nil {
		log.Println("Error decoding response to json", err)
		errorPkg.RenderInternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
