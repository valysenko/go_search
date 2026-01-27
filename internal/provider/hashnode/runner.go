package hashnode

import (
	"context"
	"errors"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"log/slog"
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
	logger            *slog.Logger
	tags              []string
	maxConcurrentTags int64
}

func NewHashnodeRunner(hashnode HashnodeProvider, logger *slog.Logger, tags []string, maxConcurrentTags int64) *HashnodeRunner {
	return &HashnodeRunner{
		hashnode:          hashnode,
		logger:            logger,
		tags:              tags,
		maxConcurrentTags: maxConcurrentTags,
	}
}

func (hr *HashnodeRunner) Run(ctx context.Context, articlesFrom time.Time) error {
	var errrs []error

	for _, tag := range hr.tags {
		err := hr.hashnode.FetchArticles(ctx, articlesFrom, provider.Query{
			TagSlug: tag,
		})
		if err != nil {
			errrs = append(errrs, fmt.Errorf("hashnode: tag %q failed: %w", tag, err))
		}
	}

	return errors.Join(errrs...)
}

func (hr *HashnodeRunner) RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	hr.logger.Info("runner started")
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(hr.maxConcurrentTags)

	for _, tag := range hr.tags {
		wg.Add(1)
		tag := tag

		go func() {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				// prevent deadlock in case errChan is blocked forever
				select {
				case errChan <- fmt.Errorf("hashnode semaphore acquire failed for tag %s: %w", tag, err):
				case <-ctx.Done():
					return
				}
				return
			}
			defer sem.Release(1)

			hr.logger.Info("fetching tag", "tag", tag)

			query := provider.Query{TagSlug: tag}
			if err := hr.hashnode.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
				// prevent deadlock in case errChan is blocked forever
				select {
				case errChan <- fmt.Errorf("hashnode runner failed for tag '%s': %w", tag, err):
				case <-ctx.Done():
					return
				}
				return

			}
			hr.logger.Info("completed tag", "tag", tag)
		}()
	}

	wg.Wait()
	hr.logger.Info("runner completed successfully")
}
