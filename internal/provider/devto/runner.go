package devto

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"log/slog"
	"time"
)

type DevToProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
	FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error
}

type DevToRunner struct {
	devto  DevToProvider
	tags   []string
	logger *slog.Logger
}

func NewDevToRunner(devto DevToProvider, logger *slog.Logger, tags []string) *DevToRunner {
	return &DevToRunner{
		devto:  devto,
		tags:   tags,
		logger: logger,
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
	dr.logger.Info("run started")
	query := provider.Query{Tags: dr.tags}

	if err := dr.devto.FetchArticlesAsync(ctx, articlesFrom, query, articlesChan); err != nil {
		errChan <- &provider.ProviderError{Provider: provider.DevTo, Err: err, Msg: "devto runner failed"}
		return
	}

	dr.logger.Info("run completed successfully")
}
