package devto

import (
	"context"
	"fmt"
	"go_search/pkg/httpclient"
	"net/url"
)

const baseUrl = "https://dev.to/api"

/**
* @deprecated
* go:generatego run github.com/vektra/mockery/v2@v2.42.0 --name=HTTPClient --case snake
 */
// need to run from the root: go run github.com/vektra/mockery/v3@v3.6.1
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
	params := url.Values{}
	params.Add("tag", request.Tag)
	params.Add("per_page", fmt.Sprintf("%d", request.PerPage))
	params.Add("page", fmt.Sprintf("%d", request.Page))

	path := fmt.Sprintf("/articles?%s", params.Encode())
	var articles []ArticleSummary
	err := dc.client.Get(ctx, path, nil, &articles)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles for tag %q: %w", request.Tag, err)
	}

	return articles, nil
}

// https://developers.forem.com/api/v1#tag/articles/operation/getArticleById
func (dc *DevToClient) GetArticleById(ctx context.Context, request *GetArticlesByIdRequest) (*Article, error) {
	url := fmt.Sprintf("/articles/%d", request.ID)
	var article Article
	err := dc.client.Get(ctx, url, nil, &article)
	if err != nil {
		return nil, fmt.Errorf("failed to get article with ID %d: %w", request.ID, err)
	}

	return &article, nil
}

// https://developers.forem.com/api/v1#tag/articles/operation/getLatestArticles
func (dc *DevToClient) GetLatestArticles(ctx context.Context, request *GetLatestArticlesRequest) ([]ArticleSummary, error) {
	params := url.Values{}
	params.Add("per_page", fmt.Sprintf("%d", request.PerPage))
	params.Add("page", fmt.Sprintf("%d", request.Page))
	path := fmt.Sprintf("/articles/latest?%s", params.Encode())

	var articles []ArticleSummary
	err := dc.client.Get(ctx, path, nil, &articles)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest articles for page %d: %w", request.Page, err)
	}

	return articles, nil
}
