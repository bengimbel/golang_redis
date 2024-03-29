package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bengimbel/go_redis_api/src/model"
	"github.com/go-redis/cache/v9"
)

type genericError error

var errorNotExist genericError = errors.New("Could not find city in redis cache")

type RedisImplementor interface {
	Insert(ctx context.Context, weather model.WeatherResponse) error
	FindByCity(ctx context.Context, city string) (model.WeatherResponse, error)
	DoesKeyExist(ctx context.Context, city string) bool
}
type RedisRepo struct {
	Cache *cache.Cache
}

// Insert city weather into redis cache.
func (rds *RedisRepo) Insert(ctx context.Context, weather model.WeatherResponse) error {
	// Save city name as key
	key := strings.ToLower(weather.City.Name)

	// Redis cache saves values for 5 minutes.
	// Return an error if there is one
	if err := rds.Cache.Set(&cache.Item{
		Key:   key,
		Value: weather,
		TTL:   10 * time.Minute,
	}); err != nil {
		return fmt.Errorf("failed to insert weather object to redis: %w", err)
	}

	return nil
}

// Get city weather from redis cache.
func (rds *RedisRepo) FindByCity(ctx context.Context, city string) (model.WeatherResponse, error) {
	// Response struct for results
	weatherModel := model.WeatherResponse{}

	// Get city weather from redis cache using the city as a key.
	if err := rds.Cache.Get(ctx, city, &weatherModel); err != nil {
		return weatherModel, errorNotExist
	}

	return weatherModel, nil
}

// Check if city is in redis cache.
func (rds *RedisRepo) DoesKeyExist(ctx context.Context, city string) bool {
	// Check cache if key exists
	return rds.Cache.Exists(ctx, city)
}
