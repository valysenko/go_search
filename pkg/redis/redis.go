package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	RedisUrl      string
	RedisPassword string
	RedisDB       int
}

type AppRedis struct {
	Client *redis.Client
}

func InitRedis(redisConfig *RedisConfig) (*AppRedis, error) {
	opt := redis.Options{
		Addr:     redisConfig.RedisUrl,
		Password: redisConfig.RedisPassword,
		DB:       redisConfig.RedisDB,
	}
	redisClient := redis.NewClient(&opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		err = redisClient.Close()
		return nil, err
	}

	return &AppRedis{
		Client: redisClient,
	}, nil
}

func (ar *AppRedis) Close() {
	ar.Client.Close()
}

func (ar *AppRedis) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return ar.Client.Ping(ctx).Err()
}
