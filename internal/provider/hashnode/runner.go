package hashnode

import (
	"context"
	"errors"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/workerpool"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type HashnodeProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
	FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error
}

/*
* Hashnode runner contains several implementations:
* - Run: sequential fetching of tags
* - RunConcurrently: concurrent fetching of tags with goroutines and semaphore for limiting concurrency
* - RunConcurrentlyWP: concurrent fetching of tags using a worker pool (tasks are prepared and sent to the pool, results are collected after all tasks are done)
* - RunConcurrentlyWPStreaming: concurrent fetching of tags using a worker pool with streaming results (tasks are prepared and sent to the pool, results are processed as they come in)
 */
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

// example of worker pool usage
func (hn *HashnodeRunner) RunConcurrentlyWP(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	hn.logger.Info("worker pool runner started")
	tasks := hn.prepareTasksForWorkerPool(articlesFrom, articlesChan)
	pool := workerpool.NewPool(tasks, int(hn.maxConcurrentTags))

	responses := pool.Run(ctx)

	for _, resp := range responses {
		if resp.Err != nil {
			select {
			case errChan <- fmt.Errorf("hashnode tag failed: %w", resp.Err):
			case <-ctx.Done():
				return
			}
		}
	}

	hn.logger.Info("runner completed successfully")
}

// example of worker pool usage with streaming results
func (hn *HashnodeRunner) RunConcurrentlyWPStreaming(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	hn.logger.Info("worker pool streaming runner started")
	tasks := hn.prepareTasksForWorkerPool(articlesFrom, articlesChan)
	pool := workerpool.NewPool(tasks, int(hn.maxConcurrentTags))

	for resp := range pool.RunStream(ctx) {
		if resp.Err != nil {
			select {
			case errChan <- fmt.Errorf("hashnode tag failed: %w", resp.Err):
			case <-ctx.Done():
				return
			}
		}
	}

	hn.logger.Info("runner completed successfully")
}

func (hn *HashnodeRunner) prepareTasksForWorkerPool(articlesFrom time.Time, articlesChan chan<- *article.Article) []*workerpool.Task[string, struct{}] {
	tasks := make([]*workerpool.Task[string, struct{}], 0, len(hn.tags))
	for i, tag := range hn.tags {
		task := workerpool.NewTask(
			func(ctx context.Context, tag string) (struct{}, error) {
				hn.logger.Info("fetching tag", "tag", tag)
				err := hn.hashnode.FetchArticlesAsync(ctx, articlesFrom, provider.Query{TagSlug: tag}, articlesChan)
				if err == nil {
					hn.logger.Info("tag completed successfully", "tag", tag)
				}
				return struct{}{}, err
			},
			tag,
			i,
		)

		tasks = append(tasks, task)
	}

	return tasks
}
