package article

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetArticleByUuid(t *testing.T) {
	ctx := context.Background()
	db := setupPostgresTestDB(t)
	t.Cleanup(func() {
		truncateArticles(t, db)
		db.Close()
	})

	testArticle, _ := NewArticle(
		"uuid-123", "Test Title", "url",
		"content", "author", SourceDevTo,
		[]string{"go", "testing"}, time.Now(),
	)
	insertTestArticle(t, db, testArticle)

	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	repo := NewArticleSearchRepository(nil, db, nullLogger)

	res, err := repo.GetArticleByUuid(ctx, testArticle.UUID)
	assert.Nil(t, err)

	assert.Equal(t, testArticle.UUID, res.UUID)
	assert.Equal(t, testArticle.Title, res.Title)
	assert.Equal(t, testArticle.Author, res.Author)
}

func TestSearchArticle(t *testing.T) {
	ctx := context.Background()
	es := setupElasticsearchTestDB(t)
	t.Cleanup(func() {
		truncateArticlesES(t, es)
		es.Close(ctx)
	})

	testArticle := &Article{
		UUID:    "uuid-java-generics",
		Title:   "Generics in java is a widely used feature, added in 2004",
		Author:  "John",
		Content: "There are generics in java",
		Tags:    []string{"java", "generics", "programming"},
	}
	insertTestArticleES(t, es, testArticle)

	testArticle = &Article{
		UUID:    "uuid-go-generics",
		Title:   "Generics in golang are new feature, added several years ago",
		Author:  "John",
		Content: "There are generics in java",
		Tags:    []string{"go", "generics", "programming"},
	}
	insertTestArticleES(t, es, testArticle)

	testArticle = &Article{ // should not appear in results
		UUID:    "uuid-travelling",
		Title:   "Travelling is a wonderful experience",
		Author:  "John",
		Content: "Nice to travel",
		Tags:    []string{"travelling"},
	}
	insertTestArticleES(t, es, testArticle)

	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	repo := NewArticleSearchRepository(es, nil, nullLogger)

	res, err := repo.SearchArticle(ctx, "generics go", 3)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, "uuid-go-generics", res[0].UUID)
	assert.Equal(t, "uuid-java-generics", res[1].UUID)
}

func TestGetArticleByUuid_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupPostgresTestDB(t)
	t.Cleanup(func() {
		db.Close()
	})

	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	repo := NewArticleSearchRepository(nil, db, nullLogger)

	_, err := repo.GetArticleByUuid(ctx, "uuid")
	assert.NotNil(t, err)
}
