package devto

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"log"
	"time"
)

type DevToProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
	FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error
}

type DevToRunner struct {
	devto DevToProvider
	tags  []string
}

func NewDevToRunner(devto DevToProvider, tags []string) *DevToRunner {
	return &DevToRunner{
		devto: devto,
		tags:  tags,
	}
}

func (dr *DevToRunner) Run(ctx context.Context, articlesFrom time.Time) error {
	query := provider.Query{Tags: dr.tags}

	err := dr.devto.FetchArticles(ctx, articlesFrom, query)
	if err != nil {
		return fmt.Errorf("devto runner failed: %w", err)
	}

	return nil
}

func (dr *DevToRunner) RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	log.Println("[info] devto runner: started")
	query := provider.Query{Tags: dr.tags}

	if err := dr.devto.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
		errChan <- fmt.Errorf("devto runner failed: %w", err)
		return
	}

	log.Println("[info] devto runner: completed successfully")
}
