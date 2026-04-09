package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	AppEnv         string `env:"APP_ENV" env-default:"loc"`
	PodName        string `env:"POD_NAME"  env-default:"pod-name-was-not-set"`
	Namespace      string `env:"NAMESPACE"  env-default:"namespace-was-not-set"`
	PushgatewayURL string `env:"PUSHGATEWAY_URL" env-default:""`
	PostgreSqlConfig
	FetcherConfig
	ProvidersConfig
	RedisConfig
}

type RedisConfig struct {
	RedisUrl      string `env:"REDIS_URL"`
	RedisPassword string `env:"REDIS_PASSWORD" env-default:""`
	RedisDB       int    `env:"REDIS_DB" env-default:"0"`
}

type FetcherConfig struct {
	DbInserterBatchSize    int `env:"DB_INSERTER_BATCH_SIZE" env-default:"10"`
	MaxConcurrentProviders int `env:"MAX_CONCURRENT_PROVIDERS" env-default:"5"`
	MaxConcurrentDbWriters int `env:"MAX_CONCURRENT_DB_WRITERS" env-default:"2"`
	ArticlesChanBatchSize  int `env:"ARTICLES_CHAN_BATCH_SIZE" env-default:"500"`
	ErrorsChanBatchSize    int `env:"ERRORS_CHAN_BATCH_SIZE" env-default:"50"`
}

type ProvidersConfig struct {
	DevToTags                    []string `env:"DEVTO_TAGS" env-default:"go,programming"`
	DevClientTimeoutSeconds      int      `env:"DEVTO_CLIENT_TIMEOUT_SECONDS" env-default:"10"`
	HashnodeTags                 []string `env:"HASHNODE_TAGS" env-default:"go,programming"`
	HashnodeClientTimeoutSeconds int      `env:"HASHNODE_CLIENT_TIMEOUT_SECONDS" env-default:"10"`
	HashnodeMaxConcurrency       int64    `env:"HASHNODE_MAX_CONCURRENCY" env-default:"3"`
	WikiCategories               []string `env:"WIKI_CATEGORIES" env-default:"technology,programming"`
	WikiClientTimeoutSeconds     int      `env:"WIKI_CLIENT_TIMEOUT_SECONDS" env-default:"10"`
	WikiMaxConcurrency           int64    `env:"WIKI_MAX_CONCURRENCY" env-default:"3"`
}

func InitConfig() *AppConfig {
	cfg := &AppConfig{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("Error reading environment variables: %v", err)
		panic(err)
	}

	return cfg
}
