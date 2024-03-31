package service_test

import (
	"context"
	"testing"

	"github.com/bengimbel/go_redis_api/internal/model"
	"github.com/bengimbel/go_redis_api/internal/repository"
	"github.com/bengimbel/go_redis_api/internal/service"
	"github.com/bengimbel/go_redis_api/pkg/httpClient"
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
	Repo       *MockRedisRepo
	HttpClient *MockHttpClient
}

func (mds *MockRedisRepo) Insert(ctx context.Context, city string, weather model.WeatherResponse) error {
	args := mds.Called(ctx, city, weather)
	return args.Error(0)
}

func (mds *MockRedisRepo) FindByCity(ctx context.Context, city string) (model.WeatherResponse, error) {
	args := mds.Called(ctx, city)
	return args.Get(0).(model.WeatherResponse), args.Error(1)
}

func (mds *MockRedisRepo) DoesKeyExist(ctx context.Context, city string) bool {
	args := mds.Called(ctx, city)
	return args.Bool(0)
}

func (mhc *MockHttpClient) MakeWeatherRequest(config *httpClient.HttpConfig, responseStruct interface{}) error {
	args := mhc.Called(config, responseStruct)
	return args.Error(0)
}

var mockRepo = &MockRedisRepo{}

var mockClient = &MockHttpClient{}

var mockService = MockService{
	Repo:       mockRepo,
	HttpClient: mockClient,
}

var mockWeatherService = service.WeatherService{
	Repo:       mockRepo,
	HttpClient: mockClient,
}

func TestFetchCoordinatesSuccess(t *testing.T) {
	expected := []model.WeatherCoordinates{
		{
			Lat: 123.123000,
			Lon: 456.456000,
		},
	}
	httpConfig := &httpClient.HttpConfig{
		Path: service.FETCH_COORDIANTES_PATH,
		Query: []httpClient.QueryParams{
			{
				Key:   service.QUERY_PARAM_Q,
				Value: "chicago",
			},
			{
				Key:   service.APP_ID_KEY,
				Value: "",
			},
		},
	}
	coordinates := []model.WeatherCoordinates{}
	mockClient.On("MakeWeatherRequest", httpConfig, &coordinates).Return(nil).Once().Return(nil).Once().Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]model.WeatherCoordinates)
		*arg = append(*arg, model.WeatherCoordinates{
			Lat: 123.123000,
			Lon: 456.456000,
		})
	})
	actual, _ := mockWeatherService.FetchCoordinates(httpConfig)

	assert.EqualValues(t, expected, actual)
}

func TestFetchWeatherSuccess(t *testing.T) {
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
	httpConfig := &httpClient.HttpConfig{
		Path: service.FETCH_WEATHER_PATH,
		Query: []httpClient.QueryParams{
			{
				Key:   service.QUERY_PARAM_LAT,
				Value: "123.123000",
			},
			{
				Key:   service.QUERY_PARAM_LON,
				Value: "456.456000",
			},
			{
				Key:   service.APP_ID_KEY,
				Value: "",
			},
		},
	}
	weather := model.WeatherResponse{}
	mockClient.On("MakeWeatherRequest", httpConfig, &weather).Return(nil).Once().Return(nil).Once().Run(func(args mock.Arguments) {
		arg := args.Get(1).(*model.WeatherResponse)
		arg.City.Name = "chicago"
		arg.List = []model.List{
			{
				Dt: 123,
			},
		}
	})
	actual, _ := mockWeatherService.FetchWeatherByCity(httpConfig)

	assert.EqualValues(t, expected, actual)
}

func TestRetrieveAndCacheWeatherAsyncSuccess(t *testing.T) {
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
	coordinateConfig := &httpClient.HttpConfig{
		Path: service.FETCH_COORDIANTES_PATH,
		Query: []httpClient.QueryParams{
			{
				Key:   service.QUERY_PARAM_Q,
				Value: "chicago",
			},
			{
				Key:   service.APP_ID_KEY,
				Value: "",
			},
		},
	}
	weatherConfig := &httpClient.HttpConfig{
		Path: service.FETCH_WEATHER_PATH,
		Query: []httpClient.QueryParams{
			{
				Key:   service.QUERY_PARAM_LAT,
				Value: "123.123000",
			},
			{
				Key:   service.QUERY_PARAM_LON,
				Value: "456.456000",
			},
			{
				Key:   service.APP_ID_KEY,
				Value: "",
			},
		},
	}
	coordinates := []model.WeatherCoordinates{}
	weather := model.WeatherResponse{}
	mockClient.On("MakeWeatherRequest", coordinateConfig, &coordinates).Return(nil).Once().Return(nil).Once().Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]model.WeatherCoordinates)
		*arg = append(*arg, model.WeatherCoordinates{
			Lat: 123.123000,
			Lon: 456.456000,
		})
	})
	mockClient.On("MakeWeatherRequest", weatherConfig, &weather).Return(nil).Once().Return(nil).Once().Run(func(args mock.Arguments) {
		arg := args.Get(1).(*model.WeatherResponse)
		arg.City.Name = "chicago"
		arg.List = []model.List{
			{
				Dt: 123,
			},
		}
	})

	mockRepo.On("Insert", ctx, "chicago", expected).Return(nil).Once()
	actual, _ := mockWeatherService.RetrieveAndCacheWeatherAsync(ctx, "chicago")

	assert.EqualValues(t, expected, actual)
}

func TestRetrieveWeatherFromCacheSuccess(t *testing.T) {
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
	mockRepo.On("FindByCity", ctx, "chicago").Return(expected, nil).Once()
	actual, _ := mockWeatherService.RetrieveWeatherFromCache(ctx, "chicago")

	assert.EqualValues(t, expected, actual)
}

func TestDoesKeyExist(t *testing.T) {
	ctx := context.Background()
	expected := true
	mockRepo.On("DoesKeyExist", ctx, "chicago").Return(true).Once()
	actual := mockWeatherService.DoesKeyExist(ctx, "chicago")

	assert.EqualValues(t, expected, actual)
}
