package devto

import (
	"context"
	"fmt"
	"go_search/pkg/httpclient"
)

const baseUrl = "https://dev.to/api"

type HTTPClient interface {
	Get(ctx context.Context, path string, headers httpclient.Headers, out any) error
}

type DevToClient struct {
	client     HTTPClient
	apiBaseUrl string
}

func NewDevToClient(timeoutSeconds int) *DevToClient {
	return &DevToClient{
		client: httpclient.NewHttpClient(timeoutSeconds, baseUrl),
	}
}

// https://developers.forem.com/api/v1#tag/articles/operation/getArticles
// https://dev.to/tags
func (dc *DevToClient) GetArticlesByTag(ctx context.Context, request *GetArticlesByTagRequest) ([]ArticleSummary, error) {

	url := fmt.Sprintf("/articles?tag=%s&per_page=%d&page=%d", request.Tag, request.PerPage, request.Page)
	var articles []ArticleSummary
	err := dc.client.Get(ctx, url, nil, articles)
	if err != nil {
		return nil, err
	}

	return articles, nil

}

// https://developers.forem.com/api/v1#tag/articles/operation/getArticleById
func (dc *DevToClient) GetArticleById(ctx context.Context, request *GetArticlesByIdRequest) (*Article, error) {
	url := fmt.Sprintf("/articles/%d", request.ID)
	var article Article
	err := dc.client.Get(ctx, url, nil, &article)
	if err != nil {
		return nil, err
	}

	return &article, nil
}

// https://developers.forem.com/api/v1#tag/articles/operation/getLatestArticles
func (dc *DevToClient) GetLatestArticles(ctx context.Context, request *GetLatestArticlesRequest) ([]ArticleSummary, error) {
	url := fmt.Sprintf("/articles/latest?per_page=%d&page=%d", request.PerPage, request.Page)

	var articles []ArticleSummary
	err := dc.client.Get(ctx, url, nil, &articles)
	if err != nil {
		return nil, err
	}

	return articles, nil

}
