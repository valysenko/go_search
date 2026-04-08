package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type HttpAppConfig struct {
	AppEnv    string `env:"APP_ENV" env-default:"loc"`
	PodName   string `env:"POD_NAME"  env-default:"pod-name-was-not-set"`
	Namespace string `env:"NAMESPACE"  env-default:"namespace-was-not-set"`
	AppPort   string `env:"APP_PORT" env-default:"8097"`
	ElasticsearchConfig
	PostgreSqlConfig
	MetricsAuth
}

type MetricsAuth struct {
	BasicAuthUsername string `env:"HTTP_BASIC_AUTH_USERNAME"  env-default:"localuser"`
	BasicAuthPassword string `env:"HTTP_BASIC_AUTH_PASSWORD"  env-default:"localpassword"`
}

type ElasticsearchConfig struct {
	Host  string `env:"ELASTICSEARCH_HOST" env-default:"http://localhost:9200"`
	Index string `env:"ELASTICSEARCH_INDEX" env-default:"articles"`
}

func InitHttpAppConfig() *HttpAppConfig {
	cfg := &HttpAppConfig{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("Error reading environment variables: %v", err)
		panic(err)
	}

	return cfg
}
