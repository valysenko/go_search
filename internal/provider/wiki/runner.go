package wiki

import (
	"context"
	"errors"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"log"
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
	var errrs []error

	for _, tag := range wr.tags {
		err := wr.wiki.FetchArticles(ctx, articlesFrom, provider.Query{
			Category: tag,
		})
		if err != nil {
			errrs = append(errrs, fmt.Errorf("wiki: category %q failed: %w", tag, err))
		}
	}

	return errors.Join(errrs...)
}

func (wr *WikiRunner) RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	log.Println("[info] wiki runner: started")
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(wr.maxConcurrentTags)

	for _, category := range wr.tags {
		wg.Add(1)
		category := category

		go func() {
			defer wg.Done()
			if err := sem.Acquire(ctx, 1); err != nil {
				// prevent deadlock in case errChan is blocked forever
				select {
				case errChan <- fmt.Errorf("wiki semaphore acquire failed for category %s: %w", category, err):
				case <-ctx.Done():
					return
				}
				return
			}
			defer sem.Release(1)

			log.Printf("[info] wiki: fetching category '%s'", category)

			query := provider.Query{Category: category}
			if err := wr.wiki.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
				// prevent deadlock in case errChan is blocked forever
				select {
				case errChan <- fmt.Errorf("wiki category %s: %w", category, err):
				case <-ctx.Done():
					return
				}
				return
			}
			log.Printf("[info] wiki: category '%s' completed", category)
		}()
	}

	wg.Wait()
	log.Println("[info] wiki runner: completed successfully")
}
