package devto

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
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
		fmt.Println("Error fetching articles:", err)
		return err
	}

	return nil
}

func (dr *DevToRunner) RunConcurrently(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error) {
	query := provider.Query{Tags: dr.tags}

	if err := dr.devto.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
		errChan <- fmt.Errorf("devTo: %w", err)
	}

	fmt.Println("devTo completed")
}
