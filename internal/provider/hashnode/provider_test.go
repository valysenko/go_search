package hashnode

import (
	"context"
	"errors"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/internal/provider/hashnode/mocks"
	"go_search/pkg/hashnode"
	hashnodeClient "go_search/pkg/hashnode"
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
	hnClient := mocks.NewMockHashnodeClient(t)

	hnProvider := &Hashnode{
		client: hnClient,
		repo:   articlesRepo,
		logger: nullLogger,
	}

	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)

	publishedAt := now.Add(-20 * time.Hour)
	expectedClientResp := &hashnodeClient.PostsByTagResponse{
		Tag: hashnodeClient.Tag{
			Name:       "go",
			Slug:       "go",
			PostsCount: 3,
			Posts: hashnodeClient.FeedPostConnection{
				Edges: []hashnodeClient.PostEdge{
					{
						Node: hashnodeClient.Post{
							ID:          "1",
							Title:       "title",
							URL:         "url",
							PublishedAt: publishedAt,
							Author:      hashnodeClient.Author{Name: "author"},
							Content:     hashnodeClient.PostContent{Text: "contents"},
						},
						Cursor: "cursor1",
					},
					{
						Node: hashnodeClient.Post{
							ID:          "2",
							Title:       "title2",
							URL:         "url2",
							PublishedAt: publishedAt,
							Author:      hashnodeClient.Author{Name: "author2"},
							Content:     hashnodeClient.PostContent{Text: "contents2"},
						},
						Cursor: "cursor2",
					},
				},
				PageInfo: hashnodeClient.PageInfo{
					HasNextPage: false,
				},
			},
		},
	}

	expectedClientRespWithCursor := &hashnodeClient.PostsByTagResponse{
		Tag: hashnodeClient.Tag{
			Name:       "go",
			Slug:       "go",
			PostsCount: 3,
			Posts: hashnodeClient.FeedPostConnection{
				Edges: []hashnodeClient.PostEdge{
					{
						Node: hashnodeClient.Post{
							ID:          "3",
							Title:       "title3",
							URL:         "url3",
							PublishedAt: publishedAt,
							Author:      hashnodeClient.Author{Name: "author3"},
							Content:     hashnodeClient.PostContent{Text: "contents3"},
						},
						Cursor: "cursor1",
					},
				},
				PageInfo: hashnodeClient.PageInfo{
					HasNextPage: true,
					EndCursor:   "cursor2",
				},
			},
		},
	}

	t.Run("successfully fetch and upsert article - one client request", func(t *testing.T) {
		request := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, nil)
		hnClient.On("GetPostsByTag", ctx, request).Return(expectedClientResp, nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "1" && arg.Title == "title"
		})).Return(nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "2" && arg.Title == "title2"
		})).Return(nil).Once()

		query := provider.Query{TagSlug: "go"}
		err := hnProvider.FetchArticles(ctx, articlesSince, query)
		assert.Nil(t, err)

		hnClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("successfully fetch and upsert article - two client requests", func(t *testing.T) {
		request1 := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, nil)
		hnClient.On("GetPostsByTag", ctx, request1).Return(expectedClientRespWithCursor, nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "3" && arg.Title == "title3"
		})).Return(nil).Once()

		after := expectedClientRespWithCursor.Tag.Posts.PageInfo.EndCursor
		request2 := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, &after)
		hnClient.On("GetPostsByTag", ctx, request2).Return(expectedClientResp, nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "1" && arg.Title == "title"
		})).Return(nil).Once()

		articlesRepo.On("UpsertArticle", ctx, mock.MatchedBy(func(arg *article.Article) bool {
			return arg.ExternalID == "2" && arg.Title == "title2"
		})).Return(nil).Once()

		query := provider.Query{TagSlug: "go"}
		err := hnProvider.FetchArticles(ctx, articlesSince, query)
		assert.Nil(t, err)

		hnClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("client returned error", func(t *testing.T) {
		request := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, nil)
		clientErr := errors.New("error")
		hnClient.On("GetPostsByTag", ctx, request).Return(nil, clientErr).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		query := provider.Query{TagSlug: "go"}
		err := hnProvider.FetchArticles(ctx, articlesSince, query)
		assert.Error(t, err)
		assert.ErrorIs(t, err, clientErr)

		hnClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})
}

func TestFetchArticlesAsync(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	articlesRepo := mocks.NewMockArticleRepository(t)
	hnClient := mocks.NewMockHashnodeClient(t)

	hnProvider := &Hashnode{
		client: hnClient,
		repo:   articlesRepo,
		logger: nullLogger,
	}

	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)

	publishedAt := now.Add(-20 * time.Hour)
	expectedClientResp := &hashnodeClient.PostsByTagResponse{
		Tag: hashnodeClient.Tag{
			Name:       "go",
			Slug:       "go",
			PostsCount: 2,
			Posts: hashnodeClient.FeedPostConnection{
				Edges: []hashnodeClient.PostEdge{
					{
						Node: hashnodeClient.Post{
							ID:          "1",
							Title:       "title",
							URL:         "url",
							PublishedAt: publishedAt,
							Author:      hashnodeClient.Author{Name: "author"},
							Content:     hashnodeClient.PostContent{Text: "contents"},
						},
						Cursor: "cursor1",
					},
					{
						Node: hashnodeClient.Post{
							ID:          "2",
							Title:       "title2",
							URL:         "url2",
							PublishedAt: publishedAt,
							Author:      hashnodeClient.Author{Name: "author2"},
							Content:     hashnodeClient.PostContent{Text: "contents2"},
						},
						Cursor: "cursor2",
					},
				},
				PageInfo: hashnodeClient.PageInfo{
					HasNextPage: false,
				},
			},
		},
	}

	expectedClientRespWithCursor := &hashnodeClient.PostsByTagResponse{
		Tag: hashnodeClient.Tag{
			Name:       "go",
			Slug:       "go",
			PostsCount: 3,
			Posts: hashnodeClient.FeedPostConnection{
				Edges: []hashnodeClient.PostEdge{
					{
						Node: hashnodeClient.Post{
							ID:          "3",
							Title:       "title3",
							URL:         "url3",
							PublishedAt: publishedAt,
							Author:      hashnodeClient.Author{Name: "author3"},
							Content:     hashnodeClient.PostContent{Text: "contents3"},
						},
						Cursor: "cursor1",
					},
				},
				PageInfo: hashnodeClient.PageInfo{
					HasNextPage: true,
					EndCursor:   "cursor2",
				},
			},
		},
	}

	t.Run("successfully fetch and stream article - one client request", func(t *testing.T) {
		request := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, nil)
		hnClient.On("GetPostsByTag", ctx, request).Return(expectedClientResp, nil).Once()

		articlesChan := make(chan *article.Article, 2)
		query := provider.Query{TagSlug: "go"}

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		var err error
		go func() {
			err = hnProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
			close(articlesChan)
		}()

		var received []*article.Article
		for art := range articlesChan {
			received = append(received, art)
		}

		assert.Nil(t, err)
		assert.Equal(t, 2, len(received))
		assert.Equal(t, "1", received[0].ExternalID)
		assert.Equal(t, "2", received[1].ExternalID)

		hnClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("successfully fetch and stream article - two client requests", func(t *testing.T) {
		request1 := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, nil)
		hnClient.On("GetPostsByTag", ctx, request1).Return(expectedClientRespWithCursor, nil).Once()

		after := expectedClientRespWithCursor.Tag.Posts.PageInfo.EndCursor
		request2 := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, &after)
		hnClient.On("GetPostsByTag", ctx, request2).Return(expectedClientResp, nil).Once()

		articlesChan := make(chan *article.Article, 3)
		query := provider.Query{TagSlug: "go"}
		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		var err error
		go func() {
			err = hnProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)
			close(articlesChan)
		}()

		var received []*article.Article
		for art := range articlesChan {
			received = append(received, art)
		}

		assert.Nil(t, err)
		assert.Equal(t, 3, len(received))
		assert.Equal(t, "3", received[0].ExternalID)
		assert.Equal(t, "1", received[1].ExternalID)
		assert.Equal(t, "2", received[2].ExternalID)

		hnClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})

	t.Run("client returned error", func(t *testing.T) {
		request := hashnode.NewGetArticlesByTagRequest("go", 10, hashnode.SortByRecent, nil)
		clientErr := errors.New("error")
		hnClient.On("GetPostsByTag", ctx, request).Return(nil, clientErr).Once()

		articlesRepo.AssertNotCalled(t, "UpsertArticle")

		query := provider.Query{TagSlug: "go"}
		articlesChan := make(chan *article.Article)
		err := hnProvider.FetchArticlesAsync(ctx, articlesSince, query, articlesChan)

		assert.Error(t, err)
		assert.ErrorIs(t, err, clientErr)

		hnClient.AssertExpectations(t)
		articlesRepo.AssertExpectations(t)
	})
}
