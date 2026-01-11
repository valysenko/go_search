package devto

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	"time"
)

type DevToProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
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
	query := provider.Query{
		Tags: dr.tags,
	}

	err := dr.devto.FetchArticles(ctx, articlesFrom, query)
	if err != nil {
		fmt.Println("Error fetching articles:", err)
		return err
	}

	return nil
}
