package devto

import (
	"context"
	"fmt"
	"go_search/helpers"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/devto"
	"strconv"

	"time"
)

type ArticleRepository interface {
	UpsertArticle(ctx context.Context, article *article.Article) error
	UpsertArticlesBatch(ctx context.Context, articles []*article.Article) error
	UpsertArticlesUnnestWithoutTags(ctx context.Context, articles []*article.Article) error
}

type DevTo struct {
	client *devto.DevToClient
	repo   ArticleRepository
}

func NewDevToProvider(client *devto.DevToClient, repo ArticleRepository) *DevTo {
	return &DevTo{
		client: client,
		repo:   repo,
	}
}

func (d *DevTo) Name() string {
	return "DevTo"
}

func (d *DevTo) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error {
	page := 1
	perPage := 30
	numArticles := 0
L:
	for {
		request := devto.NewGetLatestArticlesRequest(page, perPage)

		result, err := d.client.GetLatestArticles(ctx, request)
		if err != nil {
			// todo: log error
			fmt.Println(err)
			continue
		}

		i := 0

		articles := make([]*article.Article, 0, 30)

		for _, articleSummary := range result {
			if articlesSince.After(articleSummary.PublishedAt) {
				break L
			}

			if helpers.HasAny(articleSummary.TagList, query.Tags) {
				request := devto.NewGetArticlesByIdRequest(articleSummary.ID)
				sourceArticle, err := d.client.GetArticleById(ctx, request)
				if err != nil {
					// todo: log error
					fmt.Println(err)
					continue
				}

				article, err := article.NewArticle(
					strconv.Itoa(sourceArticle.ID),
					sourceArticle.Title,
					sourceArticle.Url,
					sourceArticle.BodyMarkdown,
					sourceArticle.User.Name,
					article.SourceDevTo,
					sourceArticle.TagList,
					sourceArticle.PublishedAt,
				)
				if err != nil {
					continue
				}

				articles = append(articles, article)

				// Temporary solution for testing. Bad performance
				// err = d.repo.UpsertArticle(ctx, article)
				// if err != nil {
				// 	fmt.Println(err)
				// 	continue
				// }

				i++
				numArticles++
			}
		}

		err = d.repo.UpsertArticlesUnnestWithoutTags(ctx, articles)

		page++
		if len(result) < perPage {
			break
		}
	}

	fmt.Println("fetched " + strconv.Itoa(numArticles) + " devto articles")

	return nil
}
