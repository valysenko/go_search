package wiki

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	"time"
)

type WikiProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
}

type WikiRunner struct {
	wiki WikiProvider
	tags []string
}

func NewWikiRunner(wiki WikiProvider, tags []string) *WikiRunner {
	return &WikiRunner{
		wiki: wiki,
		tags: tags,
	}
}

func (wr *WikiRunner) Run(ctx context.Context, articlesFrom time.Time) error {
	for _, tag := range wr.tags {
		err := wr.wiki.FetchArticles(ctx, articlesFrom, provider.Query{
			Category: tag,
		})
		if err != nil {
			fmt.Println("Error fetching articles for category", tag, ":", err)
			return err
		}
	}

	return nil
}
