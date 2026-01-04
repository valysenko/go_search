package provider

import (
	"time"
)

// interface to pass to "fetcher"
// type Provider interface {
// 	Name() string
// 	FetchArticles(ctx context.Context, articlesSince time.Time, query Query) ([]Article, error)
// }

type Source string

const (
	SourceWiki     Source = "wiki"
	SourceHashnode Source = "hashnode"
	SourceDevTo    Source = "devto"
)

type Article struct {
	ID          string
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
