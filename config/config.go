package config

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
