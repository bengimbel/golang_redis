package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bengimbel/go_redis_api/httpWeatherClient"
	"github.com/bengimbel/go_redis_api/repository"
)

type WeatherHandler struct {
	Repo              *repository.RedisRepo
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

// Handler for fetching weather from open weather map API.
func (wh *WeatherHandler) HandleFetchWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))

	// Fetch weather from open weather map API.
	// If error, log error and return general server error.
	result, err := wh.FetchWeather(city)
	if err != nil {
		fmt.Println("Error fetching city's weather", err)
		RenderInternalServerError(w, err)
		return
	}

	// Insert results into redis cache
	if err := wh.Repo.Insert(r.Context(), result); err != nil {
		fmt.Println("Error adding city weather to redis cache", err)
	}

	// Marshal struct to json for the return.
	// If error while decoding to json,
	// render a general server error
	response, err := json.Marshal(&result)
	if err != nil {
		fmt.Println("Error decoding response to json", err)
		RenderInternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (wh *WeatherHandler) HandleRetrieveCachedWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))

	// Get results from redis cache
	result, err := wh.Repo.FindByCity(r.Context(), city)
	if err != nil {
		fmt.Println("Error fetching city from redis cache", err)
		RenderInternalServerError(w, err)
		return
	}

	// Marshal struct to json for the return.
	// If error while decoding to json,
	// render a general server error
	response, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error decoding response to json", err)
		RenderInternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
