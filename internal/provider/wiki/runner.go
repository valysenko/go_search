package wiki

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type WikiProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
	FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error
}

type WikiRunner struct {
	wiki              WikiProvider
	tags              []string
	maxConcurrentTags int64
}

func NewWikiRunner(wiki WikiProvider, tags []string, maxConcurrentTags int64) *WikiRunner {
	return &WikiRunner{
		wiki:              wiki,
		tags:              tags,
		maxConcurrentTags: maxConcurrentTags,
	}
}

func (wr *WikiRunner) Run(ctx context.Context, articlesFrom time.Time) error {
	for _, tag := range wr.tags {
		err := wr.wiki.FetchArticles(ctx, articlesFrom, provider.Query{
			Category: tag,
		})
		if err != nil {
			fmt.Println("error fetching articles for category", tag, ":", err)
			return err
		}
	}

	return nil
}

func (wr *WikiRunner) RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(wr.maxConcurrentTags)

	for _, category := range wr.tags {
		wg.Add(1)
		category := category

		go func() {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				errChan <- fmt.Errorf("wiki semaphore acquire failed for category %s: %w", category, err)
				return
			}
			defer sem.Release(1)

			query := provider.Query{Category: category}
			if err := wr.wiki.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
				errChan <- fmt.Errorf("wiki category %s: %w", category, err)
			} else {
				fmt.Printf("wiki category '%s' completed\n", category)
			}
		}()
	}

	wg.Wait()
	fmt.Println("wiki completed")
}
