package hashnode

import (
	"context"
	"go_search/internal/hashnode/provider"
	"time"
)

func ExampleHashnodeProvider() {
	expectedTag := "programming"
	s := "2025-12-25T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	provider := provider.NewHashnode()
	provider.FetchArticles(context.Background(), articlesFrom, expectedTag)
}
