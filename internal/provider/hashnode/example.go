package hashnode

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	"strconv"
	"time"
)

func ExampleHashnodeProvider() {
	expectedTag := "programming"
	s := "2026-01-03T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	pr := NewHashnode()
	result, _ := pr.FetchArticles(context.Background(), articlesFrom, provider.Query{
		TagSlug: expectedTag,
	})

	i := 0
	for _, article := range result {
		fmt.Println("#" + strconv.Itoa(i) + " - " + article.PublishedAt.GoString() + " - " + article.URL + " - " + article.Author + " - " + article.Title + " - " + fmt.Sprintf("%v", article.Tags))
		i++
	}
}
