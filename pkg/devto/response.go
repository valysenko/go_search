package devto

import "time"

type ArticleSummary struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	TagList     []string  `json:"tag_list"`
	Url         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

type Article struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	BodyHTML     string    `json:"body_html"`
	BodyMarkdown string    `json:"body_markdown"`
	TagList      []string  `json:"tag_list"`
	Url          string    `json:"url"`
	PublishedAt  time.Time `json:"published_at"`
}
