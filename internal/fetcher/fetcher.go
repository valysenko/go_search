package fetcher

import (
	"context"
	"errors"
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

func (f *Fetcher) RunSequential(ctx context.Context) error {
	// temp. need to store date in redis for fetcher or per provider
	s := "2026-01-11T12:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	for _, runner := range f.providerRunners {
		err := runner.Run(ctx, articlesFrom)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (f *Fetcher) RunConcurrently(ctx context.Context) error {
	s := "2026-01-12T12:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	articlesChan := make(chan *article.Article, 100)
	errChan := make(chan error, len(f.providerRunners))

	var batchInserterWg sync.WaitGroup
	batchInserterWg.Add(1)
	go f.batchInserter(ctx, articlesChan, &batchInserterWg)

	// todo: try errgroup
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

	// 1. provders finish work
	// 2. articlesChan channel closed
	// 3. batch inserter finishes work because articlesChan is closed
	// 4. err channel closed
	go func() {
		providersWg.Wait()
		close(articlesChan)
		batchInserterWg.Wait()
		close(errChan)
	}()

	var errs []error
	for err := range errChan {
		errs = append(errs, err)

		// temp: use logger
		log.Printf("[error] provider failed: %v", err)
	}

	if len(errs) > 0 {
		log.Printf("[warn] fetch completed with %d errors", len(errs))

		return fmt.Errorf("fetch failed with %d errors: %w", len(errs), errors.Join(errs...))
	}

	log.Println("[info] All providers completed successfully")

	return nil
}

func (f *Fetcher) batchInserter(ctx context.Context, articlesChan <-chan *article.Article, wg *sync.WaitGroup) {
	defer wg.Done()

	batch := make([]*article.Article, 0, f.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := f.articleRepository.UpsertArticlesBatch(ctx, batch); err != nil {
			fmt.Printf("batch insert failed: %v\n", err)
		} else {
			fmt.Printf("inserted batch of %d articles\n", len(batch))
		}

		batch = make([]*article.Article, 0, f.batchSize)
	}

	for {
		select {
		case article, ok := <-articlesChan:
			if !ok {
				flush()
				return
			}

			batch = append(batch, article)

			if len(batch) >= f.batchSize {
				flush()
			}

		case <-ctx.Done():
			flush()
			return
		}
	}
}
