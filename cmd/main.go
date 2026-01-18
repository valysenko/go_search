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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	hashnodeRunner := hashnode.NewHashnodeRunner(hashnodeProvider, tagsFromEnv, 3)
	wikiRunner := wiki.NewWikiRunner(wikiProvider, tagsFromEnv, 3)

	f := fetcher.NewFetcher(
		articleRepository,
		10,
		3,
		devtoRunner,
		hashnodeRunner,
		wikiRunner,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// result, err := f.RunSequential(ctx)
	result, err := f.RunConcurrently(ctx)
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
