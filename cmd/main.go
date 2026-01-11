package main

import (
	"context"
	"go_search/internal/article"
	"go_search/internal/fetcher"
	"go_search/internal/provider/devto"
	"go_search/internal/provider/hashnode"
	"go_search/internal/provider/wiki"
	"go_search/pkg/database"
	devtoClient "go_search/pkg/devto"
	wikiCLientPkg "go_search/pkg/wiki"
)

func main() {
	ctx := context.Background()
	dbConfig := database.NewDBConfig("go-search-postgres", "5432", "root", "root", "go_search_db", 10, 10, 5)
	appDb := database.InitDB(dbConfig)
	appDb.RunMigrations("./migrations")
	defer appDb.Close()

	articleRepository := article.NewArticleRepository(appDb)
	devtoProvider := devto.NewDevToProvider(devtoClient.NewDevToClient(10), articleRepository)
	hashnodeProvider := hashnode.NewHashnode(articleRepository)
	wikiProvider := wiki.NewWiki(wikiCLientPkg.NewWikiClient(10), articleRepository)

	tagsFromEnv := []string{"golang", "java", "php", "physics", "programming", "ai", "technology"}
	devtoRunner := devto.NewDevToRunner(devtoProvider, tagsFromEnv)
	hashnodeRunner := hashnode.NewHashnodeRunner(hashnodeProvider, tagsFromEnv)
	wikiRunner := wiki.NewWikiRunner(wikiProvider, tagsFromEnv)

	f := fetcher.NewFetcher(
		devtoRunner,
		hashnodeRunner,
		wikiRunner,
	)
	f.RunSequential(ctx)
}
