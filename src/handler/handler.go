package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bengimbel/go_redis_api/src/httpWeatherClient"
	"github.com/bengimbel/go_redis_api/src/model"
	"github.com/bengimbel/go_redis_api/src/repository"
)

type WeatherHandler struct {
	Repo              repository.RedisImplementor
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

// Function that wraps logic to interact with
// open weather api and caches results
func (wh *WeatherHandler) RetrieveAndCacheWeather(ctx context.Context, city string) (model.WeatherResponse, error) {
	// Fetches weather
	result, err := wh.FetchWeather(city)
	if err != nil {
		fmt.Println("Error fetching city", err)
		return model.WeatherResponse{}, err
	}
	// Inserts into redis cache
	if err := wh.Repo.Insert(ctx, result); err != nil {
		fmt.Println("Error adding city weather to redis cache", err)
	}

	return result, nil
}

// Function that wraps logic to interact with
// redis cache and find results
func (wh *WeatherHandler) RetrieveWeatherByCache(ctx context.Context, city string) (model.WeatherResponse, error) {
	// Finds city's weather by key
	result, err := wh.Repo.FindByCity(ctx, city)
	if err != nil {
		fmt.Println("Error fetching city from redis cache", err)
		return model.WeatherResponse{}, err
	}
	return result, nil
}

// Handler for fetching weather from open weather map API.
func (wh *WeatherHandler) HandleRetrieveWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))
	ctx := r.Context()
	var result model.WeatherResponse

	// Check the cache before fetching
	keyExists := wh.Repo.DoesKeyExist(ctx, city)

	if keyExists {
		value, err := wh.RetrieveWeatherByCache(ctx, city)
		if err != nil {
			RenderInternalServerError(w, err)
			return
		}
		result = value
	} else {
		value, err := wh.RetrieveAndCacheWeather(ctx, city)
		if err != nil {
			RenderBadRequestError(w, err)
			return
		}
		result = value
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
	ctx := r.Context()

	// Get results from redis cache
	result, err := wh.RetrieveWeatherByCache(ctx, city)
	if err != nil {
		RenderBadRequestError(w, err)
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
