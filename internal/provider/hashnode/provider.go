package hashnode

import (
	"context"
	"errors"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/hashnode"
	"log"
	"time"
)

type articleHandler func(ctx context.Context, art *article.Article) error

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

func (hn *Hashnode) fetchArticlesBase(
	ctx context.Context,
	articlesSince time.Time,
	query provider.Query,
	handler articleHandler,
) error {
	sortBy := hashnode.SortByRecent
	first := 10
	numArticles := 0
	var after *string

L:
	for {
		request := hashnode.NewGetArticlesByTagRequest(query.TagSlug, first, sortBy, after)
		responseData, err := hn.client.GetPostsByTag(ctx, request)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("hashnode: fetch cancelled for tag '%s': %w", query.TagSlug, err)
			}
			return fmt.Errorf("hashnode: failed to get posts for tag %q: %w", query.TagSlug, err)
		}

		for _, edge := range responseData.Tag.Posts.Edges {
			if edge.Node.PublishedAt.Before(articlesSince) {
				break L
			}

			post := edge.Node
			art, err := article.NewArticle(
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
				log.Printf("[warn] hashnode: failed to create article from post %s: %v", post.ID, err)
				continue
			}

			if err := handler(ctx, art); err != nil {
				return err
			}

			numArticles++
		}

		if !responseData.Tag.Posts.PageInfo.HasNextPage {
			break L
		}
		after = &responseData.Tag.Posts.PageInfo.EndCursor
	}

	log.Printf("[info] hashnode: fetched %d articles for tag '%s'", numArticles, query.TagSlug)
	return nil
}

func (hn *Hashnode) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error {
	handler := func(ctx context.Context, art *article.Article) error {
		if err := hn.repo.UpsertArticle(ctx, art); err != nil {
			log.Printf("[warn] hashnode: failed to upsert article %s: %v", art.ExternalID, err)
		}
		return nil
	}

	return hn.fetchArticlesBase(ctx, articlesSince, query, handler)
}

func (hn *Hashnode) FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error {
	handler := func(ctx context.Context, art *article.Article) error {
		select {
		case articlesChan <- art:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("cancelled while sending article %s: %w", art.ExternalID, ctx.Err())
		}
	}

	return hn.fetchArticlesBase(ctx, articlesSince, query, handler)
}
