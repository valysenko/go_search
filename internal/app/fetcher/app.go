package fetcher

import (
	"context"
	"go_search/config"
	"go_search/internal/article"
	"go_search/internal/provider/devto"
	"go_search/internal/provider/hashnode"
	"go_search/internal/provider/wiki"
	"go_search/pkg/database"
	"go_search/pkg/redis"
	"log"
)

type FetcherApp struct {
	db    *database.AppDB
	redis *redis.AppRedis
	cfg   *config.AppConfig
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

	return &FetcherApp{
		db:    appDb,
		redis: appRedis,
		cfg:   cfg,
	}
}

func (fa *FetcherApp) Run(ctx context.Context) {
	articleRepository := article.NewArticleRepository(fa.db)
	devtoProvider := NewDevtoProvider(articleRepository, fa.cfg.ProvidersConfig.DevClientTimeoutSeconds)
	hashnodeProvider := NewHashnodeProvider(articleRepository, fa.cfg.ProvidersConfig.HashnodeClientTimeoutSeconds)
	wikiProvider := NewWikiProvider(articleRepository, fa.cfg.ProvidersConfig.WikiClientTimeoutSeconds)

	devtoRunner := devto.NewDevToRunner(devtoProvider, fa.cfg.ProvidersConfig.DevToTags)
	hashnodeRunner := hashnode.NewHashnodeRunner(hashnodeProvider, fa.cfg.ProvidersConfig.HashnodeTags, fa.cfg.ProvidersConfig.HashnodeMaxConcurrency)
	wikiRunner := wiki.NewWikiRunner(wikiProvider, fa.cfg.ProvidersConfig.WikiCategories, fa.cfg.ProvidersConfig.WikiMaxConcurrency)

	fetcherRepository := NewFetcherStorage(fa.redis)
	fetcher := NewFetcher(
		articleRepository,
		fetcherRepository,
		fa.cfg.FetcherConfig.ArticlesBarchSize,
		fa.cfg.FetcherConfig.MaxConcurrency,
		devtoRunner,
		hashnodeRunner,
		wikiRunner,
	)

	result, err := fetcher.RunSequential(ctx)
	// result, err := fetcher.RunConcurrently(ctx)

	if err != nil {
		log.Fatalf("[fatal] Fetcher could not start: %v", err)
	}
	log.Printf("[info] Fetch completed in %v", result.Duration)
	if len(result.Errors) > 0 {
		log.Printf("[warn] %d runner errors occurred during execution:", len(result.Errors))
		for _, runnerErr := range result.Errors {
			log.Printf("  - %v", runnerErr)
		}
	}
}

func (app *FetcherApp) Close(ctx context.Context) {
	app.db.Postgresql.Close()
}
