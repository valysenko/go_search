package httpfx

import (
	"context"
	"go_search/config"
	"go_search/internal/article"
	"go_search/pkg/database"
	"go_search/pkg/es"
	"log/slog"
	"os"

	"go.uber.org/fx"
)

func ProvideConfig() *config.HttpAppConfig {
	return config.InitHttpAppConfig()
}

func ProvideLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func ProvideDatabase(lc fx.Lifecycle, cfg *config.HttpAppConfig, logger *slog.Logger) *database.AppDB {
	db := database.InitDB(&database.DBConfig{
		Host:           cfg.PostgreSqlConfig.Host,
		Port:           cfg.PostgreSqlConfig.Port,
		Username:       cfg.PostgreSqlConfig.Username,
		Password:       cfg.PostgreSqlConfig.Password,
		DbName:         cfg.PostgreSqlConfig.DbName,
		MaxConns:       cfg.PostgreSqlConfig.MaxConns,
		MinConns:       cfg.PostgreSqlConfig.MinConns,
		ConnectTimeout: cfg.PostgreSqlConfig.ConnectTimeout,
	})
	db.RunMigrations("./migrations")
	logger.Info("database connected")

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database")
			db.Postgresql.Close()
			return nil
		},
	})

	return db
}

func ProvideElasticsearch(
	lc fx.Lifecycle,
	cfg *config.HttpAppConfig,
	logger *slog.Logger,
) (*es.Client, error) {
	client, err := es.NewClient(&es.ESConfig{
		Addresses: []string{cfg.ElasticsearchConfig.Host},
		Index:     cfg.ElasticsearchConfig.Index,
	})
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}
	logger.Info("elasticsearch connected")

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing elasticsearch")
			return client.Close(ctx)
		},
	})
	return client, nil
}

func ProvideArticleSearchRepository(
	esClient *es.Client,
	db *database.AppDB,
	logger *slog.Logger,
) *article.ArticleSearchRepository {
	return article.NewArticleSearchRepository(esClient, db, logger)
}
