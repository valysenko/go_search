package fetcherfx

import (
	"go_search/config"
	"go_search/internal/article"
	"go_search/internal/fetcher"
	"go_search/internal/provider/devto"
	"go_search/internal/provider/hashnode"
	"go_search/internal/provider/wiki"
	devtoClient "go_search/pkg/devto"
	wikiCLientPkg "go_search/pkg/wiki"
	"log/slog"
)

/**
* PROVIDERS
 */

func ProvideDevtoProvider(articleRepository *article.ArticleRepository, logger *slog.Logger, cfg *config.AppConfig) devto.DevToProvider {
	return devto.NewDevToProvider(
		devtoClient.NewDevToClient(cfg.ProvidersConfig.DevClientTimeoutSeconds),
		articleRepository,
		logger.With("component", "devto_provider"),
	)
}

func ProvideHashnodeProvider(articleRepository *article.ArticleRepository, logger *slog.Logger) hashnode.HashnodeProvider {
	return hashnode.NewHashnode(
		articleRepository,
		logger.With("component", "hashnode_provider"),
	)
}

func ProvideWikiProvider(articleRepository *article.ArticleRepository, logger *slog.Logger, cfg *config.AppConfig) wiki.WikiProvider {
	return wiki.NewWiki(
		wikiCLientPkg.NewWikiClient(cfg.ProvidersConfig.WikiClientTimeoutSeconds),
		logger.With("component", "wiki_provider"),
		articleRepository,
	)
}

/**
* RUNNERS
 */

func ProvideDevtoRunner(devtoProvider devto.DevToProvider, logger *slog.Logger, cfg *config.AppConfig) *devto.DevToRunner {
	return devto.NewDevToRunner(
		devtoProvider,
		logger.With("component", "devto_runner", "tags", cfg.ProvidersConfig.DevToTags),
		cfg.ProvidersConfig.DevToTags,
	)
}

func ProvideHashnodeRunner(hashnodeProvider hashnode.HashnodeProvider, logger *slog.Logger, cfg *config.AppConfig) *hashnode.HashnodeRunner {
	return hashnode.NewHashnodeRunner(
		hashnodeProvider,
		logger.With("component", "hashnode_runner", "tags", cfg.ProvidersConfig.HashnodeTags),
		cfg.ProvidersConfig.HashnodeTags,
		cfg.ProvidersConfig.HashnodeMaxConcurrency,
	)
}

func ProvideWikiRunner(wikiProvider wiki.WikiProvider, logger *slog.Logger, cfg *config.AppConfig) *wiki.WikiRunner {
	return wiki.NewWikiRunner(
		wikiProvider,
		logger.With("component", "wiki_runner", "tags", cfg.ProvidersConfig.WikiCategories),
		cfg.ProvidersConfig.WikiCategories,
		cfg.ProvidersConfig.WikiMaxConcurrency,
	)
}

/*
* FETCHER
 */
func ProvideFetcher(articleRepository *article.ArticleRepository, fetcherStorage *fetcher.Storage, batchWriter *fetcher.DbBatchWriter, logger *slog.Logger, cfg *config.AppConfig, providerRunners []fetcher.ProviderRunner) *fetcher.Fetcher {
	fetcher := fetcher.NewFetcher(
		articleRepository,
		fetcherStorage,
		batchWriter,
		logger,
		&fetcher.FetcherParams{
			MaxConcurrentProviders: cfg.FetcherConfig.MaxConcurrentProviders,
			MaxConcurrentDbWriters: cfg.FetcherConfig.MaxConcurrentDbWriters,
			ArticlesChanBatchSize:  cfg.FetcherConfig.ArticlesChanBatchSize,
			ErrorsChanBatchSize:    cfg.FetcherConfig.ErrorsChanBatchSize,
		},
		providerRunners...,
	)

	return fetcher
}
