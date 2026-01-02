package devto

import (
	"context"
	"go_search/internal/devto/provider"
	devtoClient "go_search/pkg/devto"
	"time"
)

func ExampleProvider() {
	provider := provider.NewDevToProvider(devtoClient.NewDevToClient(10))

	expectedTags := []string{"go", "golang", "programming", "ai"}
	s := "2026-01-02T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	provider.FetchArticles(context.Background(), articlesFrom, expectedTags)
}
