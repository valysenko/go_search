package hashnode

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type HashnodeProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
	FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error
}

type HashnodeRunner struct {
	hashnode          HashnodeProvider
	tags              []string
	maxConcurrentTags int64
}

func NewHashnodeRunner(hashnode HashnodeProvider, tags []string, maxConcurrentTags int64) *HashnodeRunner {
	return &HashnodeRunner{
		hashnode:          hashnode,
		tags:              tags,
		maxConcurrentTags: maxConcurrentTags,
	}
}

func (hr *HashnodeRunner) Run(ctx context.Context, articlesFrom time.Time) error {
	for _, tag := range hr.tags {
		err := hr.hashnode.FetchArticles(ctx, articlesFrom, provider.Query{
			TagSlug: tag,
		})
		if err != nil {
			fmt.Println("Error fetching articles for tag", tag, ":", err)
			return err
		}
	}

	return nil
}

func (hr *HashnodeRunner) RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(hr.maxConcurrentTags)

	for _, tag := range hr.tags {
		wg.Add(1)
		tag := tag

		go func() {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				errChan <- fmt.Errorf("hashnode semaphore acquire failed for tag %s: %w", tag, err)
				return
			}
			defer sem.Release(1)

			query := provider.Query{TagSlug: tag}
			if err := hr.hashnode.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
				errChan <- fmt.Errorf("hashnode tag %s: %w", tag, err)
			} else {
				fmt.Printf("hashnode tag '%s' completed\n", tag)
			}
		}()
	}

	wg.Wait()
	fmt.Println("hashnode completed")
}
