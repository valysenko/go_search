package fetcher

import (
	"context"
	"go_search/config"
	"go_search/internal/article"
	"go_search/internal/fetcher"
	"go_search/internal/provider/devto"
	"go_search/internal/provider/hashnode"
	"go_search/internal/provider/wiki"
	"go_search/pkg/database"
	"go_search/pkg/redis"
	"log/slog"
	"os"
)

type FetcherApp struct {
	db     *database.AppDB
	redis  *redis.AppRedis
	cfg    *config.AppConfig
	logger *slog.Logger
}

func NewFetcherApp(cfg *config.AppConfig) *FetcherApp {
	appDb := database.InitDB(&database.DBConfig{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Username:       cfg.Username,
		Password:       cfg.Password,
		DbName:         cfg.DbName,
		MaxConns:       cfg.MaxConns,
		MinConns:       cfg.MinConns,
		ConnectTimeout: cfg.ConnectTimeout,
	})
	appDb.RunMigrations("./migrations")

	appRedis := redis.InitRedis(&redis.RedisConfig{
		RedisUrl:      cfg.RedisConfig.RedisUrl,
		RedisPassword: cfg.RedisConfig.RedisPassword,
		RedisDB:       cfg.RedisConfig.RedisDB,
	})

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return &FetcherApp{
		db:     appDb,
		redis:  appRedis,
		cfg:    cfg,
		logger: logger,
	}
}

func (fa *FetcherApp) Run(ctx context.Context) {
	articleRepository := article.NewArticleRepository(fa.db)
	devtoProvider := NewDevtoProvider(articleRepository, fa.logger, fa.cfg.ProvidersConfig.DevClientTimeoutSeconds)
	hashnodeProvider := NewHashnodeProvider(articleRepository, fa.logger, fa.cfg.ProvidersConfig.HashnodeClientTimeoutSeconds)
	wikiProvider := NewWikiProvider(articleRepository, fa.logger, fa.cfg.ProvidersConfig.WikiClientTimeoutSeconds)

	devtoRunner := devto.NewDevToRunner(devtoProvider, fa.logger, fa.cfg.ProvidersConfig.DevToTags)
	hashnodeRunner := hashnode.NewHashnodeRunner(hashnodeProvider, fa.logger, fa.cfg.ProvidersConfig.HashnodeTags, fa.cfg.ProvidersConfig.HashnodeMaxConcurrency)
	wikiRunner := wiki.NewWikiRunner(wikiProvider, fa.logger, fa.cfg.ProvidersConfig.WikiCategories, fa.cfg.ProvidersConfig.WikiMaxConcurrency)

	fetcherRepository := NewFetcherStorage(fa.redis)
	batchWriter := NewDbBatchWriter(articleRepository, fa.logger, fa.cfg.FetcherConfig.DbInserterBatchSize)
	fetcher := NewFetcher(
		articleRepository,
		fetcherRepository,
		batchWriter,
		fa.logger,
		&fetcher.FetcherParams{
			MaxConcurrentProviders: fa.cfg.FetcherConfig.MaxConcurrentProviders,
			MaxConcurrentDbWriters: fa.cfg.FetcherConfig.MaxConcurrentDbWriters,
			ArticlesChanBatchSize:  fa.cfg.FetcherConfig.ArticlesChanBatchSize,
			ErrorsChanBatchSize:    fa.cfg.FetcherConfig.ErrorsChanBatchSize,
		},
		devtoRunner,
		hashnodeRunner,
		wikiRunner,
	)

	// result, err := fetcher.RunSequential(ctx)
	result, err := fetcher.RunConcurrently(ctx)

	if err != nil {
		fa.logger.Error("fetcher could not start",
			"error", err,
		)
		os.Exit(1)
	}

	fa.logger.Info("fetch completed", "duration", result.Duration.Seconds())

	if len(result.Errors) > 0 {
		fa.logger.Warn("runner errors occurred during execution",
			"error_count", len(result.Errors),
		)
		for _, runnerErr := range result.Errors {
			fa.logger.Warn("runner error", "err", runnerErr)
		}
	}
}

func (app *FetcherApp) Close(ctx context.Context) {
	app.db.Postgresql.Close()
}
