package devto

import (
	"context"
	"fmt"
	"go_search/helpers"
	"go_search/internal/provider"
	"go_search/pkg/devto"
	"strconv"

	"time"
)

type DevTo struct {
	client *devto.DevToClient
}

func NewDevToProvider(client *devto.DevToClient) *DevTo {
	return &DevTo{
		client: client,
	}
}

func (d *DevTo) Name() string {
	return "DevTo"
}

func (d *DevTo) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) ([]provider.Article, error) {
	page := 1
	perPage := 30
	response := []provider.Article{}
L:
	for {
		request := devto.NewGetLatestArticlesRequest(page, perPage)

		result, _ := d.client.GetLatestArticles(ctx, request)
		i := 0

		for _, articleSummary := range result {
			if articlesSince.After(articleSummary.PublishedAt) {
				break L
			}

			if helpers.HasAny(articleSummary.TagList, query.Tags) {
				request := devto.NewGetArticlesByIdRequest(articleSummary.ID)
				article, err := d.client.GetArticleById(ctx, request)
				if err != nil {
					// todo: log error
					fmt.Println(err)
					continue
				}

				response = append(response, provider.Article{
					ID:          strconv.Itoa(article.ID),
					Title:       article.Title,
					URL:         article.Url,
					Content:     article.BodyMarkdown,
					PublishedAt: article.PublishedAt,
					Tags:        article.TagList,
					Author:      article.User.Name,
					Source:      provider.SourceDevTo,
				})

				i++
			}

		}

		page++
		if len(result) < perPage {
			break
		}
	}
	return response, nil
}
