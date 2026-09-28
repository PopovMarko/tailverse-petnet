package core_redis

import (
	"context"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address  string        `envconfig:"ADDRESS" default:"localhost:6379"`
	Password string        `envconfig:"PASSWORD" default:""`
	DB       int           `envconfig:"DB" default:"0"`
	Timeout  time.Duration `envconfig:"TIMEOUT" default:"5s"`
}

func NewRedisConfigMust() RedisConfig {
	var config RedisConfig
	if err := envconfig.Process("REDIS", &config); err != nil {
		panic(fmt.Errorf("process redis config: %w", err))
	}
	return config
}

func NewClient(ctx context.Context, config RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.Timeout,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	})

	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis %s: %w", config.Address, err)
	}
	return client, nil
}
