package hashnode

import (
	"context"

	"github.com/machinebox/graphql"
)

const baseUrl = "https://gql.hashnode.com/"

type GQLClient interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type HashnodeClient struct {
	client GQLClient
}

func NewHashnodeClient() *HashnodeClient {
	return &HashnodeClient{
		client: graphql.NewClient(baseUrl),
	}
}

const postsByTagQuery = `
query PostsByTagQuery($slug: String!, $first: Int!, $sortBy: TagPostsSort!, $after: String) {
	tag(slug: $slug) {
		name
		slug
		postsCount
		posts(first: $first, after: $after, filter: { sortBy: $sortBy }) {
			edges {
				node {
					id
					title
					url
					author {
						name
					}
					content {
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
`

// query - https://apidocs.hashnode.com/?source=legacy-api-page#query-tag
// cursor pagination - https://apidocs.hashnode.com/?source=legacy-api-page#introduction-item-6
func (hc *HashnodeClient) GetPostsByTag(ctx context.Context, request *PostsByTagRequest) (*PostsByTagResponse, error) {
	req := graphql.NewRequest(postsByTagQuery)
	req.Var("slug", request.TagSlug)
	req.Var("first", request.First)
	req.Var("sortBy", request.SortBy)
	if request.After != nil {
		req.Var("after", *request.After)
	}

	var resp PostsByTagResponse
	if err := hc.client.Run(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
