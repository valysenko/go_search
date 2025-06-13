package client

import (
	"context"

	"github.com/machinebox/graphql"
)

type GQLClient interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type HashnodeClient struct {
	client GQLClient
}

func NewHashnodeClient() *HashnodeClient {
	return &HashnodeClient{
		client: graphql.NewClient("https://gql.hashnode.com/"),
	}
}

// query - https://apidocs.hashnode.com/?source=legacy-api-page#query-tag
// cursor pagination - https://apidocs.hashnode.com/?source=legacy-api-page#introduction-item-6
func (hc *HashnodeClient) GetPostsByTag(request *PostsByTagRequest) (*PostsByTagResponse, error) {
	req := graphql.NewRequest(`
		query PostsByTagQuery($slug: String!, $first: Int!, $sortBy: TagPostsSort!, $after: String) {
			tag(slug: $slug) {
				name
				slug
				postsCount
				posts(first: $first, after: $after, filter: { sortBy:  $sortBy }) {
					edges {
						node {
							id
							title
							url
							author {
								username
							}
							content {
							    markdown
								html
								text
							}
							publishedAt
						}
						cursor
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`)

	req.Var("slug", request.TagSlug)
	req.Var("first", request.First)
	req.Var("sortBy", request.SortBy)
	if request.After != nil {
		req.Var("after", *request.After)
	}

	var resp PostsByTagResponse
	if err := hc.client.Run(context.Background(), req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
