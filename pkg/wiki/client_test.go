package wiki

import (
	"context"
	"go_search/pkg/httpclient"
	"go_search/pkg/wiki/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCategoryMembers(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	wikiClient := WikiClient{
		client: httpClient,
	}

	ctx := context.Background()
	request := NewGetCategoryMembersRequest("golang", "")
	urlValues := request.UrlValues()
	expectedPath := "?" + urlValues.Encode()
	ecpectedHeaders := httpclient.Headers{
		userAgentHeader: userAgent,
	}

	t.Run("get category members success", func(t *testing.T) {
		expectedResponse := &CategoryMembersResponse{
			Query: struct {
				CategoryMembers []CategoryMember `json:"categorymembers"`
			}{
				CategoryMembers: []CategoryMember{
					{
						PageID: 123,
						Title:  "golang",
					},
				},
			},
			Continue: struct {
				CmContinue string `json:"cmcontinue"`
			}{
				CmContinue: "continue_token",
			},
		}

		httpClient.On("Get", ctx, expectedPath, ecpectedHeaders, mock.Anything).Run(func(args mock.Arguments) {
			out := args.Get(3).(*CategoryMembersResponse)
			*out = *expectedResponse
		}).Once().Return(nil)

		response, err := wikiClient.GetCategoryMembers(ctx, request)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, len(response.Query.CategoryMembers))
		assert.Equal(t, 123, response.Query.CategoryMembers[0].PageID)
		assert.Equal(t, "golang", response.Query.CategoryMembers[0].Title)
	})

	t.Run("get category members failure", func(t *testing.T) {
		expectedErr := &httpclient.RequestError{}

		httpClient.On("Get", ctx, expectedPath, ecpectedHeaders, mock.Anything).Once().Return(expectedErr)

		response, err := wikiClient.GetCategoryMembers(ctx, request)
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestGetArticleContent(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	wikiClient := WikiClient{
		client: httpClient,
	}

	ctx := context.Background()
	request := NewGetArticleContentRequest("pageId123")
	urlValues := request.UrlValues()
	expectedPath := "?" + urlValues.Encode()
	ecpectedHeaders := httpclient.Headers{
		userAgentHeader: userAgent,
	}

	t.Run("get article content success", func(t *testing.T) {
		expectedResponse := &ArticleResponse{
			Query: struct {
				Pages map[string]Page `json:"pages"`
			}{
				Pages: map[string]Page{
					"pageId123": {
						PageID:  123,
						Title:   "golang",
						Extract: "extract content",
					},
				},
			},
		}

		httpClient.On("Get", ctx, expectedPath, ecpectedHeaders, mock.Anything).Run(func(args mock.Arguments) {
			out := args.Get(3).(*ArticleResponse)
			*out = *expectedResponse
		}).Once().Return(nil)

		response, err := wikiClient.GetArticleContent(ctx, request)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, len(response.Query.Pages))
		assert.Equal(t, 123, response.Query.Pages["pageId123"].PageID)
		assert.Equal(t, "golang", response.Query.Pages["pageId123"].Title)
		assert.Equal(t, "extract content", response.Query.Pages["pageId123"].Extract)
	})

	t.Run("get article content failure", func(t *testing.T) {
		expectedErr := &httpclient.RequestError{}

		httpClient.On("Get", ctx, expectedPath, ecpectedHeaders, mock.Anything).Once().Return(expectedErr)

		response, err := wikiClient.GetArticleContent(ctx, request)
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, expectedErr)
	})
}
