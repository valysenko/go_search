package redis

import "github.com/redis/go-redis/v9"

type RedisConfig struct {
	RedisUrl      string
	RedisPassword string
	RedisDB       int
}

type AppRedis struct {
	Client *redis.Client
}

func InitRedis(redisConfig *RedisConfig) *AppRedis {
	opt := redis.Options{
		Addr:     redisConfig.RedisUrl,
		Password: redisConfig.RedisPassword,
		DB:       redisConfig.RedisDB,
	}
	redisClient := redis.NewClient(&opt)

	return &AppRedis{
		Client: redisClient,
	}
}

func (ar *AppRedis) Close() {
	ar.Client.Close()
}
