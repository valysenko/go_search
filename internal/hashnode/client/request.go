package client

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-TagPostsSort
type SortBy string

const (
	SortByRecent   SortBy = "recent"
	SortByPopular  SortBy = "popular"
	SortByTrending SortBy = "trending"
)

// other options in posts arguments - https://apidocs.hashnode.com/?source=legacy-api-page#definition-Tag
type PostsByTagRequest struct {
	TagSlug string
	First   int
	After   *string
	SortBy  SortBy
}

func NewGetArticlesByTagRequest(tagSlug string, first int, sortBy SortBy, after *string) *PostsByTagRequest {
	return &PostsByTagRequest{
		TagSlug: tagSlug,
		First:   first,
		SortBy:  sortBy,
		After:   after,
	}
}
