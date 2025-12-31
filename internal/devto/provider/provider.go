package provider

import (
	"context"
	"fmt"
	"go_search/helpers"
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

func (d *DevTo) FetchArticles(ctx context.Context, articlesSince time.Time, expectedTags []string) ([]devto.Article, error) {
	page := 1
	perPage := 30
L:
	for {
		request := devto.NewGetLatestArticlesRequest(page, perPage)

		result, _ := d.client.GeLatestArticles(request)
		fmt.Println("received count=" + strconv.Itoa(len(result)))

		i := 0

		for _, articleSummary := range result {
			if articlesSince.After(articleSummary.PublishedAt) {
				fmt.Println(articleSummary.PublishedAt.GoString() + " - " + articleSummary.Title)
				break L
			}

			if helpers.HasAny(articleSummary.TagList, expectedTags) {
				request := devto.NewGetArticlesByIdRequest(articleSummary.ID)
				article, _ := d.client.GetArticleById(request)
				fmt.Println("#" + strconv.Itoa(i) + " - " + article.PublishedAt.GoString() + " - " + article.Title)
				i++
			}

		}

		page++
		if len(result) < perPage {
			break
		}
	}
	return nil, nil
}
