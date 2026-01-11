package fetcher

import (
	"context"
	"fmt"
	"time"
)

type ProviderRunner interface {
	Run(ctx context.Context, articlesFrom time.Time) error
}

type Fetcher struct {
	providerRunners []ProviderRunner
}

func NewFetcher(providerRunners ...ProviderRunner) *Fetcher {
	return &Fetcher{
		providerRunners: providerRunners,
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
