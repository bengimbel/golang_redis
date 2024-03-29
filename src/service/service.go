package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bengimbel/go_redis_api/src/httpClient"
	"github.com/bengimbel/go_redis_api/src/model"
	"github.com/bengimbel/go_redis_api/src/repository"
	"github.com/redis/go-redis/v9"
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

type WeatherService struct {
	Repo       repository.RedisImplementor
	HttpClient httpClient.HttpImplementor
}

type WeatherServiceImplementor interface {
	FetchCoordinates(*httpClient.HttpConfig) ([]model.WeatherCoordinates, error)
	FetchWeatherByCity(*httpClient.HttpConfig) (model.WeatherResponse, error)
	RetrieveAndCacheWeather(context.Context, string) (model.WeatherResponse, error)
	RetrieveWeatherFromCache(context.Context, string) (model.WeatherResponse, error)
	DoesKeyExist(context.Context, string) bool
}

func NewWeatherService(rds *redis.Client) *WeatherService {
	return &WeatherService{
		Repo:       repository.NewRedisRepo(rds),
		HttpClient: httpClient.NewHttpClient(),
	}
}

// Build request struct for fetching city coordinates
func BuildLatLonRequest(city string) *httpClient.HttpConfig {
	return &httpClient.HttpConfig{
		Path: FETCH_COORDIANTES_PATH,
		Query: []httpClient.QueryParams{
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

// Build request struct for fetching city's weather from coordinate request
func BuildCityWeatherRequest(coordinates model.WeatherCoordinates) *httpClient.HttpConfig {
	return &httpClient.HttpConfig{
		Path: FETCH_WEATHER_PATH,
		Query: []httpClient.QueryParams{
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

// Fetch city coordinates using the weatherHTTPClient.
// If network error, return it. If no results, return a basic error
// If no error, return results
func (ws *WeatherService) FetchCoordinates(config *httpClient.HttpConfig) ([]model.WeatherCoordinates, error) {
	weatherCoordinates := []model.WeatherCoordinates{}

	if err := ws.HttpClient.MakeWeatherRequest(config, &weatherCoordinates); err != nil {
		return weatherCoordinates, errors.New("Error fetching coordinates. Check if API key is valid.")
	} else if len(weatherCoordinates) == 0 {
		return weatherCoordinates, fmt.Errorf("Error fetching city coordinates by name: %s", config.Query[0].Value)
	}

	return weatherCoordinates, nil
}

// Fetch city's weather using lat lon from above request
// using the weatherHTTPClient.
// If network error, return it. If no error, return results
func (ws *WeatherService) FetchWeatherByCity(config *httpClient.HttpConfig) (model.WeatherResponse, error) {
	weatherResponse := model.WeatherResponse{}

	if err := ws.HttpClient.MakeWeatherRequest(config, &weatherResponse); err != nil {
		return weatherResponse, fmt.Errorf("Error fetching city by coordinates: %s", err)
	}

	// Just saving and returning first entry in the list of results
	weatherResponse.List = []model.List{weatherResponse.List[0]}

	return weatherResponse, nil
}

// This wraps the two functions above, making sure that
// coordinate request succeeds first before moving to the
// second network request. We need to make two network requests
// because we first need to fetch city coordinates (lat lon) by city name
// then using the lat lon we can fetch the weather.
// If an error, we return the error with a empty struct.
// If no error we return the results struct with nil as error.
func (ws *WeatherService) RetrieveAndCacheWeather(ctx context.Context, city string) (model.WeatherResponse, error) {
	if coordinates, err := ws.FetchCoordinates(BuildLatLonRequest(city)); err == nil && len(coordinates) > 0 {
		if weatherResponse, err := ws.FetchWeatherByCity(BuildCityWeatherRequest(coordinates[0])); err == nil {
			// If both above requests are successful,
			// Insert result into redis cache.
			if err := ws.Repo.Insert(ctx, weatherResponse); err != nil {
				fmt.Println("Error adding city weather to redis cache", err)
			}
			return weatherResponse, nil
		} else {
			return model.WeatherResponse{}, err
		}
	} else {
		return model.WeatherResponse{}, err
	}
}

// Function that wraps logic to interact with
// redis cache and find results
func (ws *WeatherService) RetrieveWeatherFromCache(ctx context.Context, city string) (model.WeatherResponse, error) {
	// Finds city's weather by key
	weatherResponse, err := ws.Repo.FindByCity(ctx, city)
	if err != nil {
		fmt.Println("Error fetching city from redis cache", err)
		return model.WeatherResponse{}, err
	}
	return weatherResponse, nil
}

// Function that wraps logic to interact with
// redis cache and find results
func (ws *WeatherService) DoesKeyExist(ctx context.Context, city string) bool {
	// Checks if key exists
	return ws.Repo.DoesKeyExist(ctx, city)
}
