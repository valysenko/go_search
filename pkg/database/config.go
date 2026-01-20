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
