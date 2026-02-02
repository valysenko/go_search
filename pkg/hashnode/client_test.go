package hashnode

import (
	"context"
	"errors"
	"go_search/pkg/hashnode/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPostsByTag(t *testing.T) {
	ctx := context.Background()
	request := &PostsByTagRequest{
		TagSlug: "go",
		First:   2,
		SortBy:  SortByRecent,
	}

	publishedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	expectedResp := &PostsByTagResponse{
		Tag: Tag{
			Name:       "go",
			Slug:       "go",
			PostsCount: 3,
			Posts: FeedPostConnection{
				Edges: []PostEdge{
					{
						Node: Post{
							ID:          "1",
							Title:       "title",
							URL:         "url",
							PublishedAt: publishedAt,
							Author:      Author{Name: "author"},
							Content:     PostContent{Text: "contents"},
						},
						Cursor: "cursor1",
					},
				},
				PageInfo: PageInfo{
					HasNextPage: true,
					EndCursor:   "cursor1",
				},
			},
		},
	}

	t.Run("get posts by tag success", func(t *testing.T) {
		mockClient := mocks.NewMockGQLClient(t)
		hashnodeClient := HashnodeClient{
			client: mockClient,
		}

		// mock when function does not return response, but fills the response argument instead
		mockClient.On("Run", ctx, mock.AnythingOfType("*graphql.Request"), mock.Anything).
			Run(func(args mock.Arguments) {
				resp := args.Get(2).(*PostsByTagResponse)
				*resp = *expectedResp
			}).Return(nil)

		result, err := hashnodeClient.GetPostsByTag(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "go", result.Tag.Name)
		assert.Equal(t, "go", result.Tag.Slug)
		assert.Equal(t, 3, result.Tag.PostsCount)
		assert.Len(t, result.Tag.Posts.Edges, 1)
		assert.Equal(t, "title", result.Tag.Posts.Edges[0].Node.Title)
		assert.Equal(t, "author", result.Tag.Posts.Edges[0].Node.Author.Name)
		assert.True(t, result.Tag.Posts.PageInfo.HasNextPage)
		mockClient.AssertExpectations(t)
	})

	t.Run("get posts by tag error", func(t *testing.T) {
		mockClient := mocks.NewMockGQLClient(t)
		hashnodeClient := HashnodeClient{
			client: mockClient,
		}
		clientErr := errors.New("some error")
		mockClient.On("Run", ctx, mock.AnythingOfType("*graphql.Request"), mock.Anything).Return(clientErr)

		result, err := hashnodeClient.GetPostsByTag(ctx, request)

		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, clientErr)
		mockClient.AssertExpectations(t)
	})
}
