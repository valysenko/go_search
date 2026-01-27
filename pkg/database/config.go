package database

import (
	"fmt"
	"net/url"
)

type DBConfig struct {
	Host           string
	Port           string
	Username       string
	Password       string
	DbName         string
	MaxConns       int32
	MinConns       int32
	ConnectTimeout int
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
