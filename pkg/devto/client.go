package devto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type DevToClient struct {
	client     HTTPClient
	apiBaseUrl string
}

func NewDevToClient() *DevToClient {
	return &DevToClient{
		client:     &http.Client{},
		apiBaseUrl: "https://dev.to/api",
	}
}

// https://developers.forem.com/api/v1#tag/articles/operation/getArticles
// https://dev.to/tags
func (dc *DevToClient) GetArticlesByTag(request *GetArticlesByTagRequest) ([]ArticleSummary, error) {
	url := fmt.Sprintf("%s/articles?tag=%s&per_page=%d&page=%d", dc.apiBaseUrl, request.Tag, request.PerPage, request.Page)
	fmt.Println("url=", url)
	resp, err := dc.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var articles []ArticleSummary
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &articles)

	return articles, nil

}

// https://developers.forem.com/api/v1#tag/articles/operation/getArticleById
func (dc *DevToClient) GetArticleById(request *GetArticlesByIdRequest) (*Article, error) {
	url := fmt.Sprintf("%s/articles/%d", dc.apiBaseUrl, request.ID)
	resp, err := dc.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var article *Article
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &article)

	return article, nil
}

// https://developers.forem.com/api/v1#tag/articles/operation/getLatestArticles
func (dc *DevToClient) GeLatestArticles(request *GetLatestArticlesRequest) ([]ArticleSummary, error) {
	url := fmt.Sprintf("%s/articles/latest?per_page=%d&page=%d", dc.apiBaseUrl, request.PerPage, request.Page)
	fmt.Println("url=", url)
	resp, err := dc.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var articles []ArticleSummary
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &articles)

	return articles, nil

}
