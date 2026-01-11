package hashnode

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/database"
	"time"
)

func ExampleHashnodeProvider(appDB *database.AppDB, repo *article.ArticleRepository) {
	expectedTag := "programming"
	s := "2026-01-03T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	pr := NewHashnode(repo)
	err = pr.FetchArticles(context.Background(), articlesFrom, provider.Query{
		TagSlug: expectedTag,
	})
	if err != nil {
		fmt.Println("Error fetching articles:", err)
		return
	}
}
