package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bengimbel/go_redis_api/internal/model"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

type RedisImplementor interface {
	Insert(context.Context, string, model.WeatherResponse) error
	FindByCity(context.Context, string) (model.WeatherResponse, error)
	DoesKeyExist(context.Context, string) bool
}
type RedisRepo struct {
	Cache *cache.Cache
}

// Setting Cache to use local in-process storage
// to cache the small subset of recent keys.
// Key/Values use LRU (least recently used)
// for 1 minute in local in-process storage
// before looking into the Redis Cache.
func NewRedisRepo(rds *redis.Client) *RedisRepo {
	return &RedisRepo{
		Cache: cache.New(&cache.Options{
			Redis:      rds,
			LocalCache: cache.NewTinyLFU(1000, time.Minute),
		}),
	}
}

// Insert city weather into redis cache.
func (rds *RedisRepo) Insert(ctx context.Context, city string, weather model.WeatherResponse) error {
	// Save city name as key
	key := strings.ToLower(city)
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
		return weatherModel, fmt.Errorf("Could not find city in redis cache: %s", city)
	}

	return weatherModel, nil
}

// Check if city is in redis cache.
func (rds *RedisRepo) DoesKeyExist(ctx context.Context, city string) bool {
	// Check cache if key exists
	return rds.Cache.Exists(ctx, city)
}
