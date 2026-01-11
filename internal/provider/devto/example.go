package devto

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/database"
	devtoClient "go_search/pkg/devto"
	"time"
)

func ExampleProvider(appDB *database.AppDB, repo *article.ArticleRepository) {
	pr := NewDevToProvider(devtoClient.NewDevToClient(10), repo)

	expectedTags := []string{"go", "golang", "programming", "ai"}
	query := provider.Query{
		Tags: expectedTags,
	}
	s := "2026-01-10T12:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	err = pr.FetchArticles(context.Background(), articlesFrom, query)
	if err != nil {
		fmt.Println("Error fetching articles:", err)
		return
	}
}
