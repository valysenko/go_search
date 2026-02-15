package wiki

import (
	"context"
	"errors"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/internal/provider/wiki/mocks"
	"go_search/pkg/wiki"
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
	wikiClient := mocks.NewMockWikiClient(t)

	wikiProvider := &Wiki{
		client: wikiClient,
		repo:   articlesRepo,
		logger: nullLogger,
	}

	ctx := context.Background()
	articlesSince := time.Now().Add(-24 * time.Hour)
	query := provider.Query{Category: "Go"}

	categoryMembersResponse := &wiki.CategoryMembersResponse{
		Query: struct {
			CategoryMembers []wiki.CategoryMember `json:"categorymembers"`
		}{
			CategoryMembers: []wiki.CategoryMember{
				{
					PageID:    675,
					Title:     "php",
					Timestamp: articlesSince.Add(4 * time.Hour),
				},
			},
		},
		Continue: struct {
			CmContinue string `json:"cmcontinue"`
		}{
			CmContinue: "token",
		},
	}
	categoryMembersResponseWithOneWrongDateArticle := &wiki.CategoryMembersResponse{
		Query: struct {
			CategoryMembers []wiki.CategoryMember `json:"categorymembers"`
		}{
			CategoryMembers: []wiki.CategoryMember{
				{
					PageID:    123,
					Title:     "golang",
					Timestamp: articlesSince.Add(4 * time.Hour),
				},
				{
					PageID:    234,
					Title:     "java",
					Timestamp: articlesSince.Add(-4 * time.Hour),
				},
			},
		},
		Continue: struct {
			CmContinue string `json:"cmcontinue"`
		}{
			CmContinue: "token",
		},
	}
	articleResponse123 := &wiki.ArticleResponse{
		Query: struct {
			Pages map[string]wiki.Page `json:"pages"`
		}{
			Pages: map[string]wiki.Page{
				"123": {
					PageID:  123,
					Title:   "golang",
					Extract: "extract content",
				},
			},
		},
	}
	articleResponse675 := &wiki.ArticleResponse{
		Query: struct {
			Pages map[string]wiki.Page `json:"pages"`
		}{
			Pages: map[string]wiki.Page{
				"675": {
					PageID:  675,
					Title:   "php",
					Extract: "extract content",
				},
			},
		},
	}

	t.Run("successfully fetch and upsert with one members request", func(t *testing.T) {
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", ctx, request).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		contentRequest := wiki.NewGetArticleContentRequest("123")
		wikiClient.On("GetArticleContent", ctx, contentRequest).Return(articleResponse123, nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "123" && arg.Title == "golang"
		})).Return(nil).Once()

		err := wikiProvider.FetchArticles(ctx, articlesSince, query)
		assert.Nil(t, err)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("successfully fetch and upsert with two members request", func(t *testing.T) {
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		requestWithContinue := wiki.NewGetCategoryMembersRequest(query.Category, "token")
		contentRequest123 := wiki.NewGetArticleContentRequest("123")
		contentRequest675 := wiki.NewGetArticleContentRequest("675")

		wikiClient.On("GetCategoryMembers", ctx, request).Return(categoryMembersResponse, nil).Once()
		wikiClient.On("GetArticleContent", ctx, contentRequest675).Return(articleResponse675, nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "675" && arg.Title == "php"
		})).Return(nil).Once()

		wikiClient.On("GetCategoryMembers", ctx, requestWithContinue).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		wikiClient.On("GetArticleContent", ctx, contentRequest123).Return(articleResponse123, nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "123" && arg.Title == "golang"
		})).Return(nil).Once()

		err := wikiProvider.FetchArticles(ctx, articlesSince, query)
		assert.Nil(t, err)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("failed because of client error on page members request", func(t *testing.T) {
		clientErr := errors.New("error")
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", ctx, request).Return(nil, clientErr).Once()
		wikiClient.AssertNotCalled(t, "GetArticleContent")

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		err := wikiProvider.FetchArticles(ctx, articlesSince, query)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, clientErr)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("failed because of client error on concrete page request", func(t *testing.T) {
		clientErr := errors.New("error")
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", ctx, request).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		contentRequest := wiki.NewGetArticleContentRequest("123")
		wikiClient.On("GetArticleContent", ctx, contentRequest).Return(nil, clientErr).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		err := wikiProvider.FetchArticles(ctx, articlesSince, query)
		assert.Nil(t, err)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("failed because of context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())

		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", cancelCtx, request).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		contentRequest := wiki.NewGetArticleContentRequest("123")
		cancel()
		wikiClient.On("GetArticleContent", cancelCtx, contentRequest).Return(nil, context.Canceled).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		err := wikiProvider.FetchArticles(cancelCtx, articlesSince, query)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, context.Canceled)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})
}

func TestFetchArticlesAsync(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	articlesRepo := mocks.NewMockArticleRepository(t)
	wikiClient := mocks.NewMockWikiClient(t)

	wikiProvider := &Wiki{
		client: wikiClient,
		repo:   articlesRepo,
		logger: nullLogger,
	}

	ctx := context.Background()
	articlesSince := time.Now().Add(-24 * time.Hour)
	query := provider.Query{Category: "Go"}

	categoryMembersResponse := &wiki.CategoryMembersResponse{
		Query: struct {
			CategoryMembers []wiki.CategoryMember `json:"categorymembers"`
		}{
			CategoryMembers: []wiki.CategoryMember{
				{
					PageID:    675,
					Title:     "php",
					Timestamp: articlesSince.Add(4 * time.Hour),
				},
			},
		},
		Continue: struct {
			CmContinue string `json:"cmcontinue"`
		}{
			CmContinue: "token",
		},
	}
	categoryMembersResponseWithOneWrongDateArticle := &wiki.CategoryMembersResponse{
		Query: struct {
			CategoryMembers []wiki.CategoryMember `json:"categorymembers"`
		}{
			CategoryMembers: []wiki.CategoryMember{
				{
					PageID:    123,
					Title:     "golang",
					Timestamp: articlesSince.Add(4 * time.Hour),
				},
				{
					PageID:    234,
					Title:     "java",
					Timestamp: articlesSince.Add(-4 * time.Hour),
				},
			},
		},
		Continue: struct {
			CmContinue string `json:"cmcontinue"`
		}{
			CmContinue: "token",
		},
	}
	articleResponse123 := &wiki.ArticleResponse{
		Query: struct {
			Pages map[string]wiki.Page `json:"pages"`
		}{
			Pages: map[string]wiki.Page{
				"123": {
					PageID:  123,
					Title:   "golang",
					Extract: "extract content",
				},
			},
		},
	}
	articleResponse675 := &wiki.ArticleResponse{
		Query: struct {
			Pages map[string]wiki.Page `json:"pages"`
		}{
			Pages: map[string]wiki.Page{
				"675": {
					PageID:  675,
					Title:   "php",
					Extract: "extract content",
				},
			},
		},
	}

	t.Run("successfully fetch and upsert with one members request", func(t *testing.T) {
		articlesChan := make(chan *article.Article, 1)
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", ctx, request).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		contentRequest := wiki.NewGetArticleContentRequest("123")
		wikiClient.On("GetArticleContent", ctx, contentRequest).Return(articleResponse123, nil).Once()
		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		errChan := make(chan error, 1)
		go func() {
			defer close(articlesChan)
			errChan <- wikiProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
		}()

		var received []*article.Article
		for art := range articlesChan {
			received = append(received, art)
		}

		err := <-errChan
		assert.Nil(t, err)
		assert.Equal(t, 1, len(received))
		assert.Equal(t, "123", received[0].ExternalID)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("successfully fetch and upsert with two members request", func(t *testing.T) {
		articlesChan := make(chan *article.Article, 2)
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		requestWithContinue := wiki.NewGetCategoryMembersRequest(query.Category, "token")
		contentRequest123 := wiki.NewGetArticleContentRequest("123")
		contentRequest675 := wiki.NewGetArticleContentRequest("675")

		wikiClient.On("GetCategoryMembers", ctx, request).Return(categoryMembersResponse, nil).Once()
		wikiClient.On("GetArticleContent", ctx, contentRequest675).Return(articleResponse675, nil).Once()

		wikiClient.On("GetCategoryMembers", ctx, requestWithContinue).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		wikiClient.On("GetArticleContent", ctx, contentRequest123).Return(articleResponse123, nil).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		errChan := make(chan error, 1)
		go func() {
			defer close(articlesChan)
			errChan <- wikiProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
		}()

		var received []*article.Article
		for art := range articlesChan {
			received = append(received, art)
		}

		err := <-errChan
		assert.Nil(t, err)
		assert.Equal(t, 2, len(received))
		assert.Equal(t, "675", received[0].ExternalID)
		assert.Equal(t, "123", received[1].ExternalID)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("failed because of client error on page members request", func(t *testing.T) {
		clientErr := errors.New("error")
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", ctx, request).Return(nil, clientErr).Once()
		wikiClient.AssertNotCalled(t, "GetArticleContent")

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		articlesChan := make(chan *article.Article)
		err := wikiProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, clientErr)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("failed because of client error on concrete page request", func(t *testing.T) {
		clientErr := errors.New("error")
		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", ctx, request).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		contentRequest := wiki.NewGetArticleContentRequest("123")
		wikiClient.On("GetArticleContent", ctx, contentRequest).Return(nil, clientErr).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		articlesChan := make(chan *article.Article)
		err := wikiProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
		assert.Nil(t, err)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("failed because of context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())

		request := wiki.NewGetCategoryMembersRequest(query.Category, "")
		wikiClient.On("GetCategoryMembers", cancelCtx, request).Return(categoryMembersResponseWithOneWrongDateArticle, nil).Once()
		contentRequest := wiki.NewGetArticleContentRequest("123")
		cancel()
		wikiClient.On("GetArticleContent", cancelCtx, contentRequest).Return(nil, context.Canceled).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")
		articlesChan := make(chan *article.Article)
		err := wikiProvider.FetchArticlesAsync(cancelCtx, articlesSince, query, articlesChan)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, context.Canceled)

		wikiClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})
}
