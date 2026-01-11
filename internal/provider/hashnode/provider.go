package hashnode

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/hashnode"
	"strconv"
	"time"
)

type ArticleRepository interface {
	UpsertArticle(ctx context.Context, article *article.Article) error
}

type Hashnode struct {
	client *hashnode.HashnodeClient
	repo   ArticleRepository
}

func NewHashnode(repo ArticleRepository) *Hashnode {
	return &Hashnode{
		client: hashnode.NewHashnodeClient(),
		repo:   repo,
	}
}

func (hn *Hashnode) Name() string {
	return "Hashnode"
}

func (hn *Hashnode) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error {
	sortBy := hashnode.SortByRecent
	first := 10
	numArticles := 0
	var after *string

L:
	for {
		request := hashnode.NewGetArticlesByTagRequest(query.TagSlug, first, sortBy, after)
		responseData, err := hn.client.GetPostsByTag(ctx, request)
		if err != nil {
			// todo: log error
			fmt.Println(err)
			continue
		}

		for _, edge := range responseData.Tag.Posts.Edges {
			if edge.Node.PublishedAt.Before(articlesSince) {
				break L
			}

			post := edge.Node

			article, err := article.NewArticle(
				post.ID,
				post.Title,
				post.URL,
				post.Content.Text,
				post.Author.Name,
				article.SourceHashnode,
				[]string{query.TagSlug},
				post.PublishedAt,
			)
			if err != nil {
				continue
			}

			// Temporary solution for testing. Bad performance
			err = hn.repo.UpsertArticle(ctx, article)
			if err != nil {
				fmt.Println(err)
				continue
			}

			numArticles++
		}

		if responseData.Tag.Posts.PageInfo.HasNextPage == false {
			break L
		}
		after = &responseData.Tag.Posts.PageInfo.EndCursor
	}

	fmt.Println("fetched " + strconv.Itoa(numArticles) + "articles")

	return nil
}
