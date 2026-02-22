package fetcher

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"log/slog"
	"sync"
	"time"
)

type FetcherStorage interface {
	GetLastFetchTime(ctx context.Context) (time.Time, error)
	SetLastFetchTime(ctx context.Context, fetchTime time.Time) error
}

type ArticleRepository interface {
	UpsertArticlesBatch(ctx context.Context, articles []*article.Article) error
}

type BatchWriter interface {
	Run(ctx context.Context, articlesChan <-chan *article.Article, errChan chan<- error)
}

type ProviderRunner interface {
	Run(ctx context.Context, articlesFrom time.Time) error
	RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error)
}

type FetcherParams struct {
	MaxConcurrentProviders int
	MaxConcurrentDbWriters int
	ArticlesChanBatchSize  int
	ErrorsChanBatchSize    int
}

type FetcherResult struct {
	Duration time.Duration
	Errors   []error
}

type Fetcher struct {
	providerRunners   []ProviderRunner
	articleRepository ArticleRepository
	fetcherStorage    FetcherStorage
	batchWriter       BatchWriter
	logger            *slog.Logger
	params            *FetcherParams
}

func NewFetcher(articleRepository ArticleRepository, fetcherStorage FetcherStorage, batchWriter BatchWriter, logger *slog.Logger, fetcherParams *FetcherParams, providerRunners ...ProviderRunner) *Fetcher {
	return &Fetcher{
		articleRepository: articleRepository,
		fetcherStorage:    fetcherStorage,
		providerRunners:   providerRunners,
		batchWriter:       batchWriter,
		logger:            logger,
		params:            fetcherParams,
	}
}

func (f *Fetcher) RunSequential(ctx context.Context) (*FetcherResult, error) {
	var runnersErrs []error
	f.logger.Info("starting sequential fetcher...")
	startTime := time.Now()

	articlesFrom, err := f.fetcherStorage.GetLastFetchTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get last fetch time: %w", err)
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

	err = f.fetcherStorage.SetLastFetchTime(ctx, time.Now())
	if err != nil {
		f.logger.Warn("failed to update last fetch time in redis", "error", err)
	}

	return &FetcherResult{
		Duration: duration,
		Errors:   nil,
	}, nil
}

func (f *Fetcher) RunConcurrently(ctx context.Context) (*FetcherResult, error) {
	f.logger.Info("starting concurrent fetcher...")
	startTime := time.Now()

	articlesFrom, err := f.fetcherStorage.GetLastFetchTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get last fetch time: %w", err)
	}

	// buffered channels because multiple providers can send articles/errors simultaneously
	articlesChan := make(chan *article.Article, f.params.ArticlesChanBatchSize)
	errChan := make(chan error, f.params.ErrorsChanBatchSize)

	// 1. Run N batch writers for saving articles concurrently
	var batchWriterWg sync.WaitGroup
	for i := 0; i < f.params.MaxConcurrentDbWriters; i++ {
		batchWriterWg.Add(1)
		go func() {
			defer batchWriterWg.Done()
			f.batchWriter.Run(ctx, articlesChan, errChan)
		}()
	}

	// 2. Run 1 errors collector
	var errsDoneSignal = make(chan struct{})
	var runnersErrs []error

	go func() {
		defer close(errsDoneSignal)
		for err := range errChan {
			runnersErrs = append(runnersErrs, err)
		}
	}()

	// 3. Run N providers for fetching articles concurrently
	// errgroup not appropriate because it is necessary to collect all errors, not just stop on the first one
	var providersWg sync.WaitGroup
	sem := make(chan struct{}, f.params.MaxConcurrentProviders)

	for _, runner := range f.providerRunners {
		providersWg.Add(1)
		// runner := runner not needed since go 1.22
		sem <- struct{}{}
		go func() {
			defer providersWg.Done()
			defer func() { <-sem }()
			runner.RunConcurrently(ctx, articlesFrom, articlesChan, errChan)
		}()
	}

	// cleanup:
	//   1. providers finish work
	//   2. articlesChan channel closed
	//   3. batch inserter finishes work because articlesChan is closed
	//   4. err channel closed
	providersWg.Wait()
	close(articlesChan)
	batchWriterWg.Wait()
	close(errChan)
	<-errsDoneSignal

	duration := time.Since(startTime)
	f.logger.Info("all providers completed", "duration", duration)
	if len(runnersErrs) > 0 {
		return &FetcherResult{
			Duration: duration,
			Errors:   runnersErrs,
		}, nil
	}

	err = f.fetcherStorage.SetLastFetchTime(ctx, time.Now())
	if err != nil {
		f.logger.Warn("failed to update last fetch time in redis", "error", err)
	}

	return &FetcherResult{
		Duration: duration,
		Errors:   nil,
	}, nil
}
