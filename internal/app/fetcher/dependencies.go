package fetcher

import (
	"go_search/internal/article"
	"go_search/internal/fetcher"
	"go_search/internal/monitoring"
	"go_search/internal/provider/devto"
	"go_search/internal/provider/hashnode"
	"go_search/internal/provider/wiki"
	"go_search/pkg/database"
	devtoClient "go_search/pkg/devto"
	"go_search/pkg/redis"
	wikiCLientPkg "go_search/pkg/wiki"
	"log/slog"
)

/*
* FETCHER
 */
func NewFetcher(
	articleRepository *article.ArticleRepository,
	fetcherStorage *fetcher.Storage,
	batchWriter *fetcher.DbBatchWriter,
	logger *slog.Logger,
	fetcherParams *fetcher.FetcherParams,
	metricsService fetcher.FetcherMetrics,
	providerRunners ...fetcher.ProviderRunner) *fetcher.Fetcher {
	return fetcher.NewFetcher(
		articleRepository,
		fetcherStorage,
		batchWriter,
		logger,
		fetcherParams,
		metricsService,
		providerRunners...,
	)
}

func NewFetcherStorage(redis *redis.AppRedis) *fetcher.Storage {
	return fetcher.NewStorage(redis)
}

func NewMetricsService(namespace, subsystem, podName, pushGatewayUrl string) *monitoring.FetcherPrometheusMetricsService {
	return monitoring.NewFetcherPrometheusMetricsService(namespace, subsystem, podName, pushGatewayUrl)
}

func NewDbBatchWriter(articleRepository *article.ArticleRepository, metricsService fetcher.BatchWriterMetrics, logger *slog.Logger, batchSize int) *fetcher.DbBatchWriter {
	return fetcher.NewDbBatchWriter(
		articleRepository,
		logger.With("component", "db_batch_writer"),
		metricsService,
		batchSize,
	)
}

/**
* PROVIDERS
 */

func NewDevtoProvider(articleRepository *article.ArticleRepository, logger *slog.Logger, timeoutSeconds int) devto.DevToProvider {
	return devto.NewDevToProvider(
		devtoClient.NewDevToClient(timeoutSeconds),
		articleRepository,
		logger.With("component", "devto_provider"),
	)
}

func NewHashnodeProvider(articleRepository *article.ArticleRepository, logger *slog.Logger, timeoutSeconds int) hashnode.HashnodeProvider {
	return hashnode.NewHashnode(
		articleRepository,
		logger.With("component", "hashnode_provider"),
	)
}

func NewWikiProvider(articleRepository *article.ArticleRepository, logger *slog.Logger, timeoutSeconds int) wiki.WikiProvider {
	return wiki.NewWiki(
		wikiCLientPkg.NewWikiClient(timeoutSeconds),
		logger.With("component", "wiki_provider"),
		articleRepository,
	)
}

/**
* RUNNERS
 */

func NewDevtoRunner(devtoProvider devto.DevToProvider, logger *slog.Logger, tags []string) *devto.DevToRunner {
	return devto.NewDevToRunner(
		devtoProvider,
		logger.With("component", "devto_runner", "tags", tags),
		tags,
	)
}

func NewHashnodeRunner(hashnodeProvider hashnode.HashnodeProvider, logger *slog.Logger, tags []string, maxConcurrency int64) *hashnode.HashnodeRunner {
	return hashnode.NewHashnodeRunner(
		hashnodeProvider,
		logger.With("component", "hashnode_runner", "tags", tags),
		tags,
		maxConcurrency,
	)
}

func NewWikiRunner(wikiProvider wiki.WikiProvider, logger *slog.Logger, tags []string, maxConcurrency int64) *wiki.WikiRunner {
	return wiki.NewWikiRunner(
		wikiProvider,
		logger.With("component", "wiki_runner", "tags", tags),
		tags,
		maxConcurrency,
	)
}

/**
* DB REPOSITORIES
 */

func NewArticleRepository(db *database.AppDB) *article.ArticleRepository {
	return article.NewArticleRepository(db)
}
