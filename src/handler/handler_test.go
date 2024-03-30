package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengimbel/go_redis_api/src/errorPkg"
	"github.com/bengimbel/go_redis_api/src/handler"
	"github.com/bengimbel/go_redis_api/src/httpClient"
	"github.com/bengimbel/go_redis_api/src/model"
	"github.com/bengimbel/go_redis_api/src/repository"
	"github.com/go-redis/cache/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	httpClient.HttpImplementor
	mock.Mock
}

type MockRedisRepo struct {
	Cache *cache.Cache
	Repo  repository.RedisImplementor
	mock.Mock
}

type MockService struct {
	Repo              *MockRedisRepo
	WeatherHTTPClient *MockHttpClient
	mock.Mock
}

type MockWeatherHandler struct {
	Service *MockService
}

func (ms *MockService) FetchCoordinates(config *httpClient.HttpConfig) ([]model.WeatherCoordinates, error) {
	args := ms.Called(config)
	return args.Get(0).([]model.WeatherCoordinates), args.Error(1)
}
func (ms *MockService) FetchWeatherByCity(config *httpClient.HttpConfig) (model.WeatherResponse, error) {
	args := ms.Called(config)
	return args.Get(0).(model.WeatherResponse), args.Error(1)
}
func (ms *MockService) RetrieveAndCacheWeatherAsync(ctx context.Context, city string) (model.WeatherResponse, error) {
	args := ms.Called(ctx, city)
	return args.Get(0).(model.WeatherResponse), args.Error(1)
}
func (ms *MockService) RetrieveWeatherFromCache(ctx context.Context, city string) (model.WeatherResponse, error) {
	args := ms.Called(ctx, city)
	return args.Get(0).(model.WeatherResponse), args.Error(1)
}
func (ms *MockService) DoesKeyExist(ctx context.Context, city string) bool {
	args := ms.Called(ctx, city)
	return args.Bool(0)
}
func (ms *MockService) InsertToCacheAsync(ctx context.Context, city string, weatherResponse model.WeatherResponse) error {
	args := ms.Called(ctx, city, weatherResponse)
	return args.Error(0)
}

var mockService = &MockService{}

var mockWeatherHandler = handler.WeatherHandler{
	Service: mockService,
}

func TestFetchWeatherFromApiSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/weather?city=chicago", nil)
	rr := httptest.NewRecorder()
	ctx := context.Background()
	expected := model.WeatherResponse{
		City: model.City{
			Name: "chicago",
		},
		List: []model.List{
			{
				Dt: 123,
			},
		},
	}

	mockService.On("DoesKeyExist", ctx, "chicago").Return(false).Once()
	mockService.On("RetrieveAndCacheWeatherAsync", ctx, "chicago").Return(expected, nil).Once()

	handler := http.HandlerFunc(mockWeatherHandler.HandleRetrieveWeather)
	handler.ServeHTTP(rr, req)

	actual := model.WeatherResponse{}
	jsonBody, _ := io.ReadAll(rr.Body)
	json.Unmarshal(jsonBody, &actual)

	assert.EqualValues(t, expected, actual)
	assert.EqualValues(t, http.StatusOK, rr.Code)
}

func TestFetchWeatherFromApiFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/weather?city=unkowncity", nil)
	rr := httptest.NewRecorder()
	ctx := context.Background()
	errorString := "Error fetching coordinates. Check if API key is valid."
	expectedError := errors.New(errorString)
	expected := errorPkg.Error{
		Code:    400,
		Message: errorString,
	}
	emptyWeather := model.WeatherResponse{}

	mockService.On("DoesKeyExist", ctx, "unkowncity").Return(false).Once()
	mockService.On("RetrieveAndCacheWeatherAsync", ctx, "unkowncity").Return(emptyWeather, expectedError).Once()

	handler := http.HandlerFunc(mockWeatherHandler.HandleRetrieveWeather)
	handler.ServeHTTP(rr, req)

	actual := errorPkg.Error{}
	jsonBody, _ := io.ReadAll(rr.Body)
	json.Unmarshal(jsonBody, &actual)

	assert.EqualValues(t, expected, actual)
	assert.EqualValues(t, http.StatusBadRequest, rr.Code)
}
