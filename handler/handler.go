package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bengimbel/go_redis_api/httpWeatherClient"
	"github.com/bengimbel/go_redis_api/model"
	"github.com/bengimbel/go_redis_api/repository"
)

const (
	FETCH_COORDIANTES_PATH string = "/geo/1.0/direct"
	FETCH_WEATHER_PATH     string = "/data/2.5/forecast"
	QUERY_PARAM_LAT        string = "lat"
	QUERY_PARAM_LON        string = "lon"
	QUERY_PARAM_Q          string = "q"
	APP_ID_KEY             string = "appid"
	APIKEY                 string = "APIKEY"
)

type WeatherHandler struct {
	Repo              *repository.RedisRepo
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

func (wh *WeatherHandler) HandleGetWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))

	result, err := wh.FetchWeather(city)
	if err != nil {
		fmt.Println("Error fetching city's weather", err)
		RenderInternalServerError(w, err)
		return
	}

	if err := wh.Repo.Insert(r.Context(), result); err != nil {
		fmt.Println("Error adding city weather to redis cache", err)
	}

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

func (wh *WeatherHandler) HandleGetCachedWeather(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))

	result, err := wh.Repo.FindByCity(r.Context(), city)
	if err != nil {
		fmt.Println("Error fetching city from redis cache", err)
		RenderInternalServerError(w, err)
		return
	}

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

func BuildLatLonRequest(city string) *httpWeatherClient.HttpConfig {
	return &httpWeatherClient.HttpConfig{
		Path: FETCH_COORDIANTES_PATH,
		Query: []httpWeatherClient.QueryParams{
			{
				Key:   QUERY_PARAM_Q,
				Value: city,
			},
			{
				Key:   APP_ID_KEY,
				Value: os.Getenv(APIKEY),
			},
		},
	}
}

func BuildCityWeatherRequest(coordinates model.WeatherCoordinates) *httpWeatherClient.HttpConfig {
	return &httpWeatherClient.HttpConfig{
		Path: FETCH_WEATHER_PATH,
		Query: []httpWeatherClient.QueryParams{
			{
				Key:   QUERY_PARAM_LAT,
				Value: fmt.Sprintf("%f", coordinates.Lat),
			},
			{
				Key:   QUERY_PARAM_LON,
				Value: fmt.Sprintf("%f", coordinates.Lon),
			},
			{
				Key:   APP_ID_KEY,
				Value: os.Getenv(APIKEY),
			},
		},
	}
}

func (wh *WeatherHandler) FetchCoordinates(config *httpWeatherClient.HttpConfig) ([]model.WeatherCoordinates, error) {
	weatherCoordinates := []model.WeatherCoordinates{}
	if err := wh.WeatherHTTPClient.MakeWeatherRequest(config, &weatherCoordinates); err != nil {
		return weatherCoordinates, fmt.Errorf("Error fetching city coordinates by name: %w", err)
	} else if len(weatherCoordinates) == 0 {
		return weatherCoordinates, fmt.Errorf("Error fetching city coordinates by name: %s", config.Query[0].Value)
	}
	return weatherCoordinates, nil
}

func (wh *WeatherHandler) FetchWeatherByCity(config *httpWeatherClient.HttpConfig) (model.WeatherResponse, error) {
	weatherResponse := model.WeatherResponse{}
	if err := wh.WeatherHTTPClient.MakeWeatherRequest(config, &weatherResponse); err != nil {
		return weatherResponse, fmt.Errorf("Error fetching city by coordinates: %s", err)
	}
	return weatherResponse, nil
}

func (wh *WeatherHandler) FetchWeather(city string) (model.WeatherResponse, error) {
	if coordinates, err := wh.FetchCoordinates(BuildLatLonRequest(city)); err == nil && len(coordinates) > 0 {
		if weatherResponse, err := wh.FetchWeatherByCity(BuildCityWeatherRequest(coordinates[0])); err == nil {
			return weatherResponse, nil
		} else {
			return model.WeatherResponse{}, err
		}
	} else {
		return model.WeatherResponse{}, err
	}
}
