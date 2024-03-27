package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bengimbel/go_redis_api/model"
	"github.com/redis/go-redis/v9"
)

type genericError error

var errorNotExist genericError = errors.New("Could not find city in redis cache")

type RedisRepo struct {
	Client *redis.Client
}

func (rds *RedisRepo) Insert(ctx context.Context, weather model.WeatherResponse) error {
	key := strings.ToLower(weather.City.Name)

	data, err := json.Marshal(weather)
	if err != nil {
		return fmt.Errorf("failed to encode weather object: %w", err)
	}

	result := rds.Client.Set(ctx, key, string(data), time.Hour)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to insert weather object to redis: %w", err)
	}

	return nil
}

func (rds *RedisRepo) FindByCity(ctx context.Context, city string) (model.WeatherResponse, error) {
	weatherModel := model.WeatherResponse{}

	value, err := rds.Client.Get(ctx, city).Result()
	if errors.Is(err, redis.Nil) {
		return weatherModel, errorNotExist
	} else if err != nil {
		return weatherModel, fmt.Errorf("Get weather error: %w", err)
	}

	unmarshalErr := json.Unmarshal([]byte(value), &weatherModel)
	if unmarshalErr != nil {
		return weatherModel, fmt.Errorf("Failed to unmarshal weather model: %w", err)
	}

	return weatherModel, nil
}
