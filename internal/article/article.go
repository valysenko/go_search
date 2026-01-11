package article

import (
	"time"

	"github.com/google/uuid"
)

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
	UpdateddAt  time.Time
	PublishedAt time.Time
	Title       string
	URL         string
	Content     string
	Author      string
	Tags        []string
	Source      Source
}

func NewArticle(externalID, title, url, content, author string, source Source, tags []string, publishedAt time.Time) (*Article, error) {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Article{
		UUID:        uuidv7.String(),
		ExternalID:  externalID,
		CreatedAt:   time.Now().UTC(),
		UpdateddAt:  time.Now().UTC(),
		Title:       title,
		URL:         url,
		Content:     content,
		Author:      author,
		Tags:        tags,
		Source:      source,
		PublishedAt: publishedAt,
	}, nil
}
