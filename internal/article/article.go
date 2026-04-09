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

func (s Source) String() string {
	switch s {
	case SourceDevTo:
		return "devto"
	case SourceHashnode:
		return "hashnode"
	case SourceWiki:
		return "wiki"
	default:
		return "unknown"
	}
}

type Article struct {
	UUID        string    `json:"uuid"`
	ExternalID  string    `json:"external_id,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdateddAt  time.Time `json:"updated_at,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	Source      Source    `json:"source"`
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

type IndexArticle struct {
	UUID        string   `json:"uuid"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Highlight   string   `json:"highlight"`
	PublishedAt string   `json:"published_at"`
}
