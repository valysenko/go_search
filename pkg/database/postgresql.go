package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

type AppDB struct {
	Postgresql *pgxpool.Pool
}

func InitDB(dbConfig *DBConfig) (*AppDB, error) {
	connStr := dbConfig.ProvideDSN()
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	poolConfig.MinConns = int32(dbConfig.MinConns)
	ctx, _ := context.WithCancel(context.Background())
	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return &AppDB{
		Postgresql: pool,
	}, nil
}

func (db *AppDB) Close() {
	db.Postgresql.Close()
}

func (db *AppDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return db.Postgresql.Ping(ctx)
}

func (db *AppDB) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return db.Postgresql.Begin(ctx)
}

func (db *AppDB) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	return db.Postgresql.BeginFunc(ctx, fn)
}

// library: go install github.com/pressly/goose/v3/cmd/goose@latest
// create migration: goose create create_article_table sql ||  goose create create_article_table go
// cli run migration:
//
//	export GOOSE_DRIVER=postgres
//	export GOOSE_DBSTRING=postgresql://root:root@go-search-postgres:5432/go_search_db?sslmode=disable
//	goose -dir ./migrations up
func (db *AppDB) RunMigrations(path string) {
	sqlDB := stdlib.OpenDB(*db.Postgresql.Config().ConnConfig)
	defer sqlDB.Close()

	goose.SetDialect("postgres")
	err := goose.Up(sqlDB, path)
	if err != nil {
		panic(err)
	}
}

func (db *AppDB) DownMigrations(path string) {
	sqlDB := stdlib.OpenDB(*db.Postgresql.Config().ConnConfig)
	defer sqlDB.Close()

	goose.SetDialect("postgres")
	err := goose.DownTo(sqlDB, path, 0)
	if err != nil {
		panic(err)
	}
}
