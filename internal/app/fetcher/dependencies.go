package fetcher

import (
	"go_search/internal/article"
	"go_search/internal/fetcher"
	"go_search/internal/provider/devto"
	"go_search/internal/provider/hashnode"
	"go_search/internal/provider/wiki"
	"go_search/pkg/database"
	devtoClient "go_search/pkg/devto"
	wikiCLientPkg "go_search/pkg/wiki"
)

/*
* FETCHER
 */
func NewFetcher(articleRepository *article.ArticleRepository, batchSize int, maxConcurrency int, providerRunners ...fetcher.ProviderRunner) *fetcher.Fetcher {
	return fetcher.NewFetcher(
		articleRepository,
		batchSize,
		maxConcurrency,
		providerRunners...,
	)
}

/**
* PROVIDERS
 */

func NewDevtoProvider(articleRepository *article.ArticleRepository, timeoutSeconds int) devto.DevToProvider {
	return devto.NewDevToProvider(devtoClient.NewDevToClient(timeoutSeconds), articleRepository)
}

func NewHashnodeProvider(articleRepository *article.ArticleRepository, timeoutSeconds int) hashnode.HashnodeProvider {
	return hashnode.NewHashnode(articleRepository)
}

func NewWikiProvider(articleRepository *article.ArticleRepository, timeoutSeconds int) wiki.WikiProvider {
	return wiki.NewWiki(wikiCLientPkg.NewWikiClient(timeoutSeconds), articleRepository)
}

/**
* RUNNERS
 */

func NewDevtoRunner(devtoProvider devto.DevToProvider, tags []string) *devto.DevToRunner {
	return devto.NewDevToRunner(devtoProvider, tags)
}

func NewHashnodeRunner(hashnodeProvider hashnode.HashnodeProvider, tags []string, maxConcurrency int64) *hashnode.HashnodeRunner {
	return hashnode.NewHashnodeRunner(hashnodeProvider, tags, maxConcurrency)
}

func NewWikiRunner(wikiProvider wiki.WikiProvider, tags []string, maxConcurrency int64) *wiki.WikiRunner {
	return wiki.NewWikiRunner(wikiProvider, tags, maxConcurrency)
}

/**
* REPOSITORIES
 */

func NewArticleRepository(db *database.AppDB) *article.ArticleRepository {
	return article.NewArticleRepository(db)
}
