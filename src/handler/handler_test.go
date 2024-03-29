package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengimbel/go_redis_api/src/handler"
	"github.com/bengimbel/go_redis_api/src/httpWeatherClient"
	"github.com/bengimbel/go_redis_api/src/model"
	"github.com/go-redis/cache/v9"
	"github.com/stretchr/testify/assert"
)

type MockWeatherHandler struct {
	Repo              *MockRedisRepo
	WeatherHTTPClient *httpWeatherClient.HttpWeatherClient
}

type MockRedisRepo struct {
	Cache *cache.Cache
}

func (mds *MockRedisRepo) Insert(ctx context.Context, weather model.WeatherResponse) error {
	return nil
}

func (mds *MockRedisRepo) FindByCity(ctx context.Context, city string) (model.WeatherResponse, error) {
	return model.WeatherResponse{
		City: model.City{
			Name: "Chicago",
		},
		List: []model.List{
			{
				Dt: 12345,
			},
		},
	}, nil
}

func (mds *MockRedisRepo) DoesKeyExist(ctx context.Context, city string) bool {
	return true
}

var mockWeatherHandler = handler.WeatherHandler{
	Repo:              &MockRedisRepo{},
	WeatherHTTPClient: httpWeatherClient.NewHttpClient(),
}

func TestFetchWeatherFromApiSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/weather?city=chicago", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockWeatherHandler.HandleRetrieveWeather)
	handler.ServeHTTP(rr, req)
	expected := model.WeatherResponse{
		City: model.City{
			Name: "Chicago",
		},
		List: []model.List{
			{
				Dt: 12345,
			},
		},
	}
	actual := model.WeatherResponse{}
	jsonBody, _ := io.ReadAll(rr.Body)
	err := json.Unmarshal(jsonBody, &actual)
	if err != nil {
		fmt.Println("Error decoding response to json", err)
	}

	assert.EqualValues(t, expected, actual)
	assert.EqualValues(t, rr.Code, http.StatusOK)
}

func TestFetchWeatherFromApiBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/weather?city=unknowncity", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockWeatherHandler.HandleRetrieveWeather)
	handler.ServeHTTP(rr, req)
	// expected := model.WeatherResponse{
	// 	// City: model.City{
	// 	// 	Name: "Chicago",
	// 	// },
	// 	// List: []model.List{
	// 	// 	{
	// 	// 		Dt: 12345,
	// 	// 	},
	// 	// },
	// }
	actual := model.WeatherResponse{}
	jsonBody, _ := io.ReadAll(rr.Body)
	err := json.Unmarshal(jsonBody, &actual)
	if err != nil {
		fmt.Println("Error decoding response to json", err)
	}

	// assert.EqualValues(t, expected, actual)
	assert.EqualValues(t, rr.Code, http.StatusBadGateway)
}
