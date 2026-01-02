package devto

import (
	"context"
	"errors"
	"testing"

	"go_search/pkg/devto/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetArticlesByTag(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	devtoClient := DevToClient{
		client: httpClient,
	}

	ctx := context.Background()
	request := &GetArticlesByTagRequest{
		Tag:     "go",
		PerPage: 10,
		Page:    1,
	}

	t.Run("get articles by tag success", func(t *testing.T) {
		expectedArticles := []ArticleSummary{
			{
				ID:    1,
				Title: "Article 1",
			},
			{
				ID:    2,
				Title: "Article 2",
			},
		}

		httpClient.On("Get", ctx, "/articles?tag=go&per_page=10&page=1", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			out := args.Get(3).(*[]ArticleSummary)
			*out = expectedArticles
		}).Once().Return(nil)

		articles, err := devtoClient.GetArticlesByTag(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, expectedArticles, articles)
	})

	t.Run("get articles by tag failure", func(t *testing.T) {
		expectedError := errors.New("request error")
		httpClient.On("Get", ctx, "/articles?tag=go&per_page=10&page=1", mock.Anything, mock.Anything).Once().Return(expectedError)

		articles, err := devtoClient.GetArticlesByTag(ctx, request)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, articles)
	})
}

func TestGetArticleById(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	devtoClient := DevToClient{
		client: httpClient,
	}

	ctx := context.Background()
	request := &GetArticlesByIdRequest{
		ID: 12,
	}

	t.Run("get article by ID success", func(t *testing.T) {
		expectedArticle := Article{
			ID:    12,
			Title: "Test Article",
		}

		httpClient.On("Get", ctx, "/articles/12", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			out := args.Get(3).(*Article)
			*out = expectedArticle
		}).Once().Return(nil)

		article, err := devtoClient.GetArticleById(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, expectedArticle.Title, article.Title)
	})

	t.Run("get article by ID failure", func(t *testing.T) {
		expectedError := errors.New("request error")
		httpClient.On("Get", ctx, "/articles/12", mock.Anything, mock.Anything).Once().Return(expectedError)

		article, err := devtoClient.GetArticleById(ctx, request)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, article)
	})
}

func TestGetLatestArticles(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	devtoClient := DevToClient{
		client: httpClient,
	}

	ctx := context.Background()
	request := &GetLatestArticlesRequest{
		PerPage: 10,
		Page:    1,
	}

	t.Run("get latest articles success", func(t *testing.T) {
		expectedArticles := []ArticleSummary{
			{
				ID:    1,
				Title: "Article 10",
			},
			{
				ID:    3,
				Title: "Article 30",
			},
		}

		httpClient.On("Get", ctx, "/articles/latest?per_page=10&page=1", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			out := args.Get(3).(*[]ArticleSummary)
			*out = expectedArticles
		}).Once().Return(nil)

		articles, err := devtoClient.GetLatestArticles(ctx, request)
		assert.NoError(t, err)
		assert.Equal(t, expectedArticles, articles)
	})

	t.Run("get latest articles error", func(t *testing.T) {
		expectedError := errors.New("request error")
		httpClient.On("Get", ctx, "/articles/latest?per_page=10&page=1", mock.Anything, mock.Anything).Once().Return(expectedError)

		articles, err := devtoClient.GetLatestArticles(ctx, request)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, articles)
	})
}
