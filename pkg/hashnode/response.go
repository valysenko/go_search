package hashnode

import "time"

type PostsByTagResponse struct {
	Tag Tag `json:"tag"`
}

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-Tag
type Tag struct {
	Name       string             `json:"name"`
	Slug       string             `json:"slug"`
	PostsCount int                `json:"postsCount"`
	Posts      FeedPostConnection `json:"posts"`
}

// // https://apidocs.hashnode.com/?source=legacy-api-page#definition-FeedPostConnection
type FeedPostConnection struct {
	Edges    []PostEdge `json:"edges"`
	PageInfo PageInfo   `json:"pageInfo"`
}

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-PostEdge
type PostEdge struct {
	Node   Post   `json:"node"`
	Cursor string `json:"cursor"`
}

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-Post
type Post struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	PublishedAt time.Time   `json:"publishedAt"`
	Author      Author      `json:"author"`
	Content     PostContent `json:"content"`
}

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-User
type Author struct {
	Username string `json:"username"`
}

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-Content
type PostContent struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Text     string `json:"text"`
}

// https://apidocs.hashnode.com/?source=legacy-api-page#definition-PageInfo
type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}
