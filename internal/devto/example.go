package devto

import (
	"go_search/internal/devto/provider"
	devtoClient "go_search/pkg/devto"
	"time"
)

func ExampleProvider() {
	provider := provider.NewDevToProvider(devtoClient.NewDevToClient())

	expectedTags := []string{"go", "golang", "programming", "ai"}
	s := "2025-12-25T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	provider.FetchArticles(nil, articlesFrom, expectedTags)
}
