package devto

import (
	"context"
	"errors"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/internal/provider/devto/mocks"
	"go_search/pkg/devto"
	"go_search/pkg/httpclient"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchArticles(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	articlesRepo := mocks.NewMockArticleRepository(t)
	devtoClient := mocks.NewMockClient(t)

	devtoProvider := &DevTo{
		client: devtoClient,
		repo:   articlesRepo,
		logger: nullLogger,
	}

	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)

	t.Run("successfully fetch and upsert article", func(t *testing.T) {
		articleSummaries := []devto.ArticleSummary{
			{ID: 123, Title: "Test Article", PublishedAt: now, TagList: []string{"go"}},
		}
		req := devto.NewGetLatestArticlesRequest(1, 30)
		devtoClient.On("GetLatestArticles", ctx, req).Return(articleSummaries, nil).Once()

		art := &devto.Article{
			ID:          123,
			Title:       "Test Article",
			PublishedAt: now,
			TagList:     []string{"go"},
			User:        devto.User{Name: "John Doe"},
			Url:         "https://dev.to/test",
		}
		req2 := devto.NewGetArticlesByIdRequest(123)
		devtoClient.On("GetArticleById", ctx, req2).Return(art, nil).Once()

		articlesRepo.On("UpsertArticlesUnnestWithoutTags", ctx, mock.MatchedBy(func(args []*article.Article) bool {
			return len(args) == 1 && args[0].ExternalID == "123" && args[0].Title == "Test Article"
		})).Return(nil).Once()

		query := provider.Query{Tags: []string{"go"}}
		err := devtoProvider.FetchArticles(ctx, articlesSince, query)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		devtoClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("test get article by id returned an error", func(t *testing.T) {
		articleSummaries := []devto.ArticleSummary{
			{ID: 123, Title: "test", PublishedAt: now, TagList: []string{"go"}},
		}
		devtoClient.On("GetLatestArticles", ctx, mock.Anything).Return(articleSummaries, nil).Once()
		devtoClient.On("GetArticleById", ctx, mock.Anything).Return(nil, fmt.Errorf("API error")).Once()
		articlesRepo.AssertNotCalled(t, "UpsertArticlesUnnestWithoutTags")

		err := devtoProvider.FetchArticles(ctx, articlesSince, provider.Query{Tags: []string{"go"}})

		assert.Nil(t, err)
	})

	t.Run("successfully fetch N matching articless", func(t *testing.T) {
		ids := []int{101, 102, 103}
		var summaries []devto.ArticleSummary

		for _, id := range ids {
			summaries = append(summaries, devto.ArticleSummary{
				ID:          id,
				Title:       fmt.Sprintf("Go Article %d", id),
				PublishedAt: now,
				TagList:     []string{"go"},
			})

			detail := &devto.Article{
				ID:          id,
				Title:       fmt.Sprintf("Go Article %d", id),
				Url:         fmt.Sprintf("https://dev.to/go/%d", id),
				PublishedAt: now,
				TagList:     []string{"go"},
				User:        devto.User{Name: "Gopher"},
			}

			devtoClient.On("GetArticleById", ctx, mock.MatchedBy(func(req *devto.GetArticlesByIdRequest) bool {
				return req.ID == id
			})).Return(detail, nil).Once()
		}

		devtoClient.On("GetLatestArticles", ctx, mock.Anything).Return(summaries, nil).Once()
		articlesRepo.On("UpsertArticlesUnnestWithoutTags", ctx, mock.MatchedBy(func(args []*article.Article) bool {
			return len(args) == len(ids)
		})).Return(nil).Once()

		query := provider.Query{Tags: []string{"go"}}
		err := devtoProvider.FetchArticles(ctx, articlesSince, query)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		devtoClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})
}

func TestFetchArticlesAsync(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	articlesRepo := mocks.NewMockArticleRepository(t)
	devtoClient := mocks.NewMockClient(t)
	devtoProvider := NewDevToProvider(devtoClient, articlesRepo, nullLogger)

	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)

	t.Run("successfully stream articles to channel", func(t *testing.T) {
		articleSummaries := []devto.ArticleSummary{
			{ID: 123, Title: "test", PublishedAt: now, TagList: []string{"go"}},
		}
		devtoClient.On("GetLatestArticles", ctx, mock.Anything).Return(articleSummaries, nil).Once()

		articleDetail := &devto.Article{
			ID:          123,
			Title:       "test",
			PublishedAt: now,
			TagList:     []string{"go"},
			User:        devto.User{Name: "test"},
			Url:         "https://dev.to/async",
		}
		devtoClient.On("GetArticleById", ctx, mock.MatchedBy(func(req *devto.GetArticlesByIdRequest) bool {
			return req.ID == 123
		})).Return(articleDetail, nil).Once()

		articlesChan := make(chan *article.Article, 1)
		query := provider.Query{Tags: []string{"go"}}

		articlesRepo.AssertNotCalled(t, "UpsertArticlesUnnestWithoutTags")

		var err error
		go func() {
			err = devtoProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
			close(articlesChan)
		}()

		var received []*article.Article
		for art := range articlesChan {
			received = append(received, art)
		}

		assert.Nil(t, err)
		assert.Equal(t, 1, len(received))
		assert.Equal(t, "123", received[0].ExternalID)

		devtoClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("stream second article to channel", func(t *testing.T) {
		articleSummaries := []devto.ArticleSummary{
			{ID: 123, Title: "test", PublishedAt: now, TagList: []string{"go"}},
			{ID: 124, Title: "test2", PublishedAt: now, TagList: []string{"go"}},
		}
		devtoClient.On("GetLatestArticles", ctx, mock.Anything).Return(articleSummaries, nil).Once()

		articleDetail2 := &devto.Article{
			ID:          124,
			Title:       "test",
			PublishedAt: now,
			TagList:     []string{"go"},
			User:        devto.User{Name: "test"},
			Url:         "https://dev.to/async",
		}
		devtoClient.On("GetArticleById", ctx, mock.MatchedBy(func(req *devto.GetArticlesByIdRequest) bool {
			return req.ID == 123
		})).Return(nil, errors.New("errrr")).Once()
		devtoClient.On("GetArticleById", ctx, mock.MatchedBy(func(req *devto.GetArticlesByIdRequest) bool {
			return req.ID == 124
		})).Return(articleDetail2, nil).Once()

		articlesChan := make(chan *article.Article, 10)
		query := provider.Query{Tags: []string{"go"}}

		var err error
		go func() {
			err = devtoProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
			close(articlesChan)
		}()

		var received []*article.Article
		for art := range articlesChan {
			received = append(received, art)
		}

		assert.Nil(t, err)
		assert.Equal(t, 1, len(received))
		assert.Equal(t, "124", received[0].ExternalID)

		devtoClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("test get article summaary returned an 500 error", func(t *testing.T) {
		devtoClient.On("GetLatestArticles", ctx, mock.Anything).Return(nil, &httpclient.RequestError{Type: httpclient.ErrorTypeServer}).Once()
		devtoClient.AssertNotCalled(t, "GetArticleById")
		articlesRepo.AssertNotCalled(t, "UpsertArticlesUnnestWithoutTags")

		articlesChan := make(chan *article.Article)
		err := devtoProvider.FetchArticlesAsync(ctx, articlesSince, provider.Query{Tags: []string{"go"}}, articlesChan)

		assert.NotNil(t, err)
	})

	t.Run("test context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())

		articleSummaries := []devto.ArticleSummary{
			{ID: 123, Title: "test", PublishedAt: now, TagList: []string{"go"}},
		}
		devtoClient.On("GetLatestArticles", cancelCtx, mock.Anything).Return(articleSummaries, nil).Once()

		// context is cancelled after fetching article summaries
		cancel()

		// this query should already return context.Canceled error
		devtoClient.On("GetArticleById", cancelCtx, mock.Anything).Return(nil, context.Canceled).Once()

		articlesChan := make(chan *article.Article, 1)
		query := provider.Query{Tags: []string{"go"}}

		err := devtoProvider.FetchArticlesAsync(cancelCtx, articlesSince, query, articlesChan)

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
