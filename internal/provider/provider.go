package provider

import (
	"time"
)

// type Provider interface {
// 	Name() string
// 	FetchArticles(ctx context.Context, articlesSince time.Time, query Query) (error)
// }

type Source int

const (
	SourceUnknown Source = iota
	SourceDevTo
	SourceHashnode
	SourceWiki
)

type Article struct {
	UUID        string
	ExternalID  string
	CreatedAt   time.Time
	Title       string
	URL         string
	Content     string
	Author      string
	PublishedAt time.Time
	Tags        []string
	Source      Source
}

type Query struct {
	Tags     []string
	Category string
	TagSlug  string
}
