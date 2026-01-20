package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	AppEnv  string `env:"APP_ENV" env-default:"loc"`
	PodName string `env:"POD_NAME"  env-default:"pod-name-was-not-set"`
	PostgreSqlConfig
	FetcherConfig
	ProvidersConfig
}

type PostgreSqlConfig struct {
	Host           string `env:"DB_HOST"`
	Port           string `env:"DB_PORT"`
	Username       string `env:"DB_USERNAME" `
	Password       string `env:"DB_PASSWORD"`
	DbName         string `env:"DB_NAME"`
	MaxConns       int32  `env:"DB_MAX_CONNS"`
	MinConns       int32  `env:"DB_MIN_CONNS"`
	ConnectTimeout int    `env:"DB_CONNECT_TIMEOUT"`
}

type FetcherConfig struct {
	ArticlesBarchSize int `env:"ARTICLES_BATCH_SIZE" env-default:"10"`
	MaxConcurrency    int `env:"MAX_CONCURRENCY" env-default:"3"`
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
		help, _ := cleanenv.GetDescription(&cfg, nil)
		fmt.Println(help)
		panic(err)
	}

	return cfg
}
