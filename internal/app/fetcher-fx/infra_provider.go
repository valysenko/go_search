package fetcherfx

import (
	"context"
	"fmt"
	"go_search/config"
	"go_search/internal/article"
	"go_search/internal/fetcher"
	"go_search/internal/monitoring"
	"go_search/pkg/database"
	"go_search/pkg/redis"
	"log"
	"log/slog"
	"os"

	"go.uber.org/fx"
)

func ProvideConfig() *config.AppConfig {
	return config.InitConfig()
}

func ProvideLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func ProvideDatabase(lc fx.Lifecycle, cfg *config.AppConfig, logger *slog.Logger) *database.AppDB {
	db, err := database.InitDB(&database.DBConfig{
		Host:           cfg.PostgreSqlConfig.Host,
		Port:           cfg.PostgreSqlConfig.Port,
		Username:       cfg.PostgreSqlConfig.Username,
		Password:       cfg.PostgreSqlConfig.Password,
		DbName:         cfg.PostgreSqlConfig.DbName,
		MaxConns:       cfg.PostgreSqlConfig.MaxConns,
		MinConns:       cfg.PostgreSqlConfig.MinConns,
		ConnectTimeout: cfg.PostgreSqlConfig.ConnectTimeout,
	})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
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

func ProvideRedis(lc fx.Lifecycle, cfg *config.AppConfig, logger *slog.Logger) *redis.AppRedis {
	appRedis, err := redis.InitRedis(&redis.RedisConfig{
		RedisUrl:      cfg.RedisConfig.RedisUrl,
		RedisPassword: cfg.RedisConfig.RedisPassword,
		RedisDB:       cfg.RedisConfig.RedisDB,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connect to redis failed: %v\n", err)
		os.Exit(1)
	}

	logger.Info("redis connected")

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing redis")
			appRedis.Close()
			return nil
		},
	})

	return appRedis
}

func ProvideArticleRepository(
	db *database.AppDB,
) *article.ArticleRepository {
	return article.NewArticleRepository(db)
}

func ProvideDbBatchWriter(articleRepository *article.ArticleRepository, logger *slog.Logger, cfg *config.AppConfig, metricsService *monitoring.FetcherPrometheusMetricsService) *fetcher.DbBatchWriter {
	return fetcher.NewDbBatchWriter(
		articleRepository,
		logger.With("component", "db_batch_writer"),
		metricsService,
		cfg.FetcherConfig.DbInserterBatchSize,
	)
}

func ProvideFetcherStorage(redis *redis.AppRedis) *fetcher.Storage {
	return fetcher.NewStorage(redis)
}

func ProvideFetcherMetrics(cfg *config.AppConfig) *monitoring.FetcherPrometheusMetricsService {
	return monitoring.NewFetcherPrometheusMetricsService(cfg.Namespace, "fx_fetcher", cfg.PodName, cfg.PushgatewayURL)
}
