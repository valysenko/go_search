package database

import (
	"fmt"
	"net/url"
)

type DBConfig struct {
	Host           string `env:"DB_HOST"`
	Port           string `env:"DB_PORT"`
	Username       string `env:"DB_USERNAME" `
	Password       string `env:"DB_PASSWORD"`
	DbName         string `env:"DB_NAME"`
	MaxConns       int32  `env:"DB_MAX_CONNS"`
	MinConns       int32  `env:"DB_MIN_CONNS"`
	ConnectTimeout int    `env:"DB_CONNECT_TIMEOUT"`
}

func NewDBConfig(
	host string,
	port string,
	username string,
	password string,
	dbName string,
	maxConns int32,
	minConns int32,
	connectTimeout int,
) *DBConfig {
	return &DBConfig{
		Host:           host,
		Port:           port,
		Username:       username,
		Password:       password,
		DbName:         dbName,
		MaxConns:       maxConns,
		MinConns:       minConns,
		ConnectTimeout: connectTimeout,
	}
}

func (cfg *DBConfig) ProvideDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=%d",
		url.QueryEscape(cfg.Username),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.DbName,
		cfg.ConnectTimeout)
}
