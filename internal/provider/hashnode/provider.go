package hashnode

import (
	"context"
	"go_search/internal/provider"
	"go_search/pkg/hashnode"
	"time"
)

type Hashnode struct {
	client *hashnode.HashnodeClient
}

func NewHashnode() *Hashnode {
	return &Hashnode{
		client: hashnode.NewHashnodeClient(),
	}
}

func (hn *Hashnode) Name() string {
	return "Hashnode"
}

func (hn *Hashnode) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) ([]provider.Article, error) {
	sortBy := hashnode.SortByRecent
	first := 10
	var after *string
	response := []provider.Article{}

L:
	for {
		request := hashnode.NewGetArticlesByTagRequest(query.TagSlug, first, sortBy, after)
		responseData, err := hn.client.GetPostsByTag(ctx, request)
		if err != nil {
			panic(err)
		}

		for _, edge := range responseData.Tag.Posts.Edges {
			if edge.Node.PublishedAt.Before(articlesSince) {
				break L
			}

			post := edge.Node

			response = append(response, provider.Article{
				ID:          post.ID,
				Title:       post.Title,
				URL:         post.URL,
				Content:     post.Content.Text,
				PublishedAt: post.PublishedAt,
				Author:      post.Author.Name,
				Tags:        []string{query.TagSlug},
				Source:      provider.SourceHashnode,
			})
		}

		if responseData.Tag.Posts.PageInfo.HasNextPage == false {
			break L
		}
		after = &responseData.Tag.Posts.PageInfo.EndCursor
	}

	return response, nil
}
