package fetcher

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"log"
	"sync"
	"time"
)

type ArticleRepository interface {
	UpsertArticlesBatch(ctx context.Context, articles []*article.Article) error
}

type ProviderRunner interface {
	Run(ctx context.Context, articlesFrom time.Time) error
	RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error)
}

type FetcherResult struct {
	Duration time.Duration
	Errors   []error
}

type Fetcher struct {
	providerRunners   []ProviderRunner
	articleRepository ArticleRepository
	batchSize         int
	maxConcurrency    int
}

func NewFetcher(articleRepository ArticleRepository, batchSize int, maxConcurrency int, providerRunners ...ProviderRunner) *Fetcher {
	return &Fetcher{
		articleRepository: articleRepository,
		providerRunners:   providerRunners,
		batchSize:         batchSize,
		maxConcurrency:    maxConcurrency,
	}
}

func (f *Fetcher) RunSequential(ctx context.Context) (*FetcherResult, error) {
	var runnersErrs []error
	log.Println("[info] starting sequential fetcher...")
	startTime := time.Now()

	// temp. need to store date in redis for fetcher or per provider
	s := "2026-01-18T08:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	for _, runner := range f.providerRunners {
		err := runner.Run(ctx, articlesFrom)
		if err != nil {
			runnersErrs = append(runnersErrs, err)
		}
	}

	duration := time.Since(startTime)

	if len(runnersErrs) > 0 {
		return &FetcherResult{
			Duration: duration,
			Errors:   runnersErrs,
		}, nil
	}

	return &FetcherResult{
		Duration: duration,
		Errors:   nil,
	}, nil
}

func (f *Fetcher) RunConcurrently(ctx context.Context) (*FetcherResult, error) {
	log.Println("[info] starting concurrent fetcher...")
	startTime := time.Now()

	// temp. need to store date in redis for fetcher or per provider
	s := "2026-01-18T08:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	// buffered channels because multiple providers can send articles/errors simultaneously
	articlesChan := make(chan *article.Article, 200)
	errChan := make(chan error, 100)

	// collect and save fetched articles
	var batchInserterWg sync.WaitGroup
	batchInserterWg.Add(1)
	go f.batchInserter(ctx, articlesChan, &batchInserterWg)

	// collect errors from runners
	var runnersErrs []error
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for err := range errChan {
			runnersErrs = append(runnersErrs, err)
		}
	}()

	// run providers. todo: try errgroup
	var providersWg sync.WaitGroup
	sem := make(chan struct{}, f.maxConcurrency)

	for _, runner := range f.providerRunners {
		providersWg.Add(1)
		runner := runner
		go func() {
			defer providersWg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			runner.RunConcurrently(ctx, articlesFrom, articlesChan, errChan)
		}()
	}

	// cleanup:
	//   1. provders finish work
	//   2. articlesChan channel closed
	//   3. batch inserter finishes work because articlesChan is closed
	//   4. err channel closed
	go func() {
		providersWg.Wait()
		close(articlesChan)
		batchInserterWg.Wait()
		close(errChan)
	}()

	collectorWg.Wait()

	duration := time.Since(startTime)
	log.Printf("[info] all providers completed in %v", duration)
	if len(runnersErrs) > 0 {
		return &FetcherResult{
			Duration: duration,
			Errors:   runnersErrs,
		}, nil
	}

	return &FetcherResult{
		Duration: duration,
		Errors:   nil,
	}, nil
}

func (f *Fetcher) batchInserter(ctx context.Context, articlesChan <-chan *article.Article, wg *sync.WaitGroup) {
	defer wg.Done()

	batch := make([]*article.Article, 0, f.batchSize)
	flush := func(flushCtx context.Context) {
		if len(batch) == 0 {
			return
		}

		if err := f.articleRepository.UpsertArticlesBatch(flushCtx, batch); err != nil {
			log.Printf("[error] batch insert failed: %v", err)
		} else {
			log.Printf("[info] inserted batch of %d articles", len(batch))
		}

		batch = make([]*article.Article, 0, f.batchSize)
	}

	for {
		select {
		case article, ok := <-articlesChan:
			if !ok {
				// pass new context for the last batch for db provider succeessful insert
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				flush(cleanupCtx)
				cancel()
				return
			}

			batch = append(batch, article)

			if len(batch) >= f.batchSize {
				flush(ctx)
			}

		case <-ctx.Done():
			// pass new context for the last batch for db provider succeessful insert
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			flush(cleanupCtx)
			cancel()
			return
		}
	}
}
