package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bengimbel/go_redis_api/httpWeatherClient"
	"github.com/bengimbel/go_redis_api/model"
	"github.com/bengimbel/go_redis_api/repository"
)

type WeatherHandler struct {
	Repo              *repository.RedisRepo
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

func (wh *WeatherHandler) HandleGetWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city := strings.ToLower(r.URL.Query().Get("city"))
	httpConfig := ConfigureLatLonRequest(city)

	result, err := wh.FetchWeather(httpConfig)
	if err != nil {
		fmt.Println("Error fetching city's weather", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := wh.Repo.Insert(r.Context(), result); err != nil {
		fmt.Println("Error adding city weather to redis cache", err)
	}

	response, err := json.Marshal(&result)
	if err != nil {
		fmt.Println("Error decoding response to json", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (wh *WeatherHandler) HandleGetCachedWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city := strings.ToLower(r.URL.Query().Get("city"))

	result, err := wh.Repo.FindByCity(r.Context(), city)
	if err != nil {
		fmt.Println("Error fetching city from redis cache", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error decoding response to json", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func ConfigureLatLonRequest(city string) *httpWeatherClient.HttpConfig {
	return &httpWeatherClient.HttpConfig{
		Path: "/geo/1.0/direct",
		Query: []httpWeatherClient.QueryParams{
			{
				Key:   "q",
				Value: city,
			},
			{
				Key:   "appid",
				Value: "<API KEY HERE>",
			},
		},
	}
}

func ConfigureCityCoordinatesRequest(coordinates model.WeatherCoordinates) *httpWeatherClient.HttpConfig {
	return &httpWeatherClient.HttpConfig{
		Path: "/data/2.5/forecast",
		Query: []httpWeatherClient.QueryParams{
			{
				Key:   "lat",
				Value: fmt.Sprintf("%f", coordinates.Lat),
			},
			{
				Key:   "lon",
				Value: fmt.Sprintf("%f", coordinates.Lon),
			},
			{
				Key:   "appid",
				Value: "<API KEY HERE>",
			},
		},
	}
}

func (wh *WeatherHandler) FetchWeather(config *httpWeatherClient.HttpConfig) (model.WeatherResponse, error) {
	weatherCoordinates := []model.WeatherCoordinates{}
	weatherResponse := model.WeatherResponse{}
	if err := wh.WeatherHTTPClient.MakeWeatherRequest(config, &weatherCoordinates); err != nil || len(weatherCoordinates) == 0 {
		return weatherResponse, fmt.Errorf("Error fetching city coordinates by name: %w", err)
	}

	httpWeatherRequestConfig := ConfigureCityCoordinatesRequest(weatherCoordinates[0])
	if err := wh.WeatherHTTPClient.MakeWeatherRequest(httpWeatherRequestConfig, &weatherResponse); err != nil {
		return weatherResponse, fmt.Errorf("Error fetching city by coordinates: %s", err)
	}
	return weatherResponse, nil
}
