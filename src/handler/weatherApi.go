package handler

import (
	"fmt"
	"os"

	"github.com/bengimbel/go_redis_api/src/httpWeatherClient"
	"github.com/bengimbel/go_redis_api/src/model"
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

// Build request struct for fetching city coordinates
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

// Build request struct for fetching city's weather from coordinate request
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

// Fetch city coordinates using the weatherHTTPClient.
// If network error, return it. If no results, return a basic error
// If no error, return results
func (wh *WeatherHandler) FetchCoordinates(config *httpWeatherClient.HttpConfig) ([]model.WeatherCoordinates, error) {
	weatherCoordinates := []model.WeatherCoordinates{}

	if err := wh.WeatherHTTPClient.MakeWeatherRequest(config, &weatherCoordinates); err != nil {
		return weatherCoordinates, fmt.Errorf("Error fetching city coordinates by name: %w", err)
	} else if len(weatherCoordinates) == 0 {
		return weatherCoordinates, fmt.Errorf("Error fetching city coordinates by name: %s", config.Query[0].Value)
	}
	return weatherCoordinates, nil
}

// Fetch city's weather using lat lon from above request
// using the weatherHTTPClient.
// If network error, return it. If no error, return results
func (wh *WeatherHandler) FetchWeatherByCity(config *httpWeatherClient.HttpConfig) (model.WeatherResponse, error) {
	weatherResponse := model.WeatherResponse{}

	if err := wh.WeatherHTTPClient.MakeWeatherRequest(config, &weatherResponse); err != nil {
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
