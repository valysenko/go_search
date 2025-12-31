package hashnode

import (
	"go_search/internal/hashnode/provider"
	hnClient "go_search/pkg/hashnode"
	"time"
)

func ExampleHashnodeProvider() {
	expectedTag := "programming"
	s := "2025-12-25T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	provider := provider.NewHashnode(hnClient.NewHashnodeClient())
	provider.FetchArticles(nil, articlesFrom, expectedTag)
}
