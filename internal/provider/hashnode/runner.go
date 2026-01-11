package hashnode

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	"time"
)

type HashnodeProvider interface {
	FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error
}

type HashnodeRunner struct {
	hashnode HashnodeProvider
	tags     []string
}

func NewHashnodeRunner(hashnode HashnodeProvider, tags []string) *HashnodeRunner {
	return &HashnodeRunner{
		hashnode: hashnode,
		tags:     tags,
	}
}

func (hr *HashnodeRunner) Run(ctx context.Context, articlesFrom time.Time) error {
	for _, tag := range hr.tags {
		err := hr.hashnode.FetchArticles(ctx, articlesFrom, provider.Query{
			TagSlug: tag,
		})
		if err != nil {
			fmt.Println("Error fetching articles for tag", tag, ":", err)
			return err
		}
	}

	return nil
}
