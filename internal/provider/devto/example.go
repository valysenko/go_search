package devto

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	devtoClient "go_search/pkg/devto"
	"strconv"
	"time"
)

func ExampleProvider() {
	pr := NewDevToProvider(devtoClient.NewDevToClient(10))

	expectedTags := []string{"go", "golang", "programming", "ai"}
	query := provider.Query{
		Tags: expectedTags,
	}
	s := "2026-01-03T15:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	result, err := pr.FetchArticles(context.Background(), articlesFrom, query)

	i := 0
	for _, article := range result {
		fmt.Println("#" + strconv.Itoa(i) + " - " + article.PublishedAt.GoString() + " - " + article.URL + " - " + article.Author + " - " + article.Title + " - " + fmt.Sprintf("%v", article.Tags))
		i++
	}
}
