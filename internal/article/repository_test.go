package article

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpsertArticle_Update(t *testing.T) {
	ctx := context.Background()
	db := setupPostgresTestDB(t)
	t.Cleanup(func() {
		truncateArticles(t, db)
		db.Close()
	})
	repo := NewArticleRepository(db)

	article, _ := NewArticle(
		"uuid-234", "Test Title", "url",
		"content", "author", SourceDevTo,
		[]string{"go", "testing"}, time.Now(),
	)
	insertTestArticle(t, db, article)

	articleToUpdate, _ := NewArticle(
		"uuid-234", "Test Title Updated", "url",
		"content updated", "author", SourceDevTo,
		[]string{"go", "testing"}, time.Now(),
	)
	err := repo.UpsertArticle(ctx, articleToUpdate)
	assert.Nil(t, err)

	updatedArticle := getTestArticle(t, db, article.ExternalID)
	assert.Equal(t, articleToUpdate.UUID, updatedArticle.UUID)
	assert.Equal(t, articleToUpdate.Title, updatedArticle.Title)
	assert.Equal(t, articleToUpdate.Author, updatedArticle.Author)
}

func TestUpsertArticle_Insert(t *testing.T) {
	ctx := context.Background()
	db := setupPostgresTestDB(t)
	t.Cleanup(func() {
		truncateArticles(t, db)
		db.Close()
	})
	repo := NewArticleRepository(db)

	article, _ := NewArticle(
		"uuid-234", "Test Title Updated", "url",
		"content updated", "author", SourceDevTo,
		[]string{"go", "testing"}, time.Now(),
	)
	err := repo.UpsertArticle(ctx, article)
	assert.Nil(t, err)

	insertedArticle := getTestArticle(t, db, article.ExternalID)

	assert.Equal(t, article.UUID, insertedArticle.UUID)
	assert.Equal(t, article.Title, insertedArticle.Title)
	assert.Equal(t, article.Author, insertedArticle.Author)
}

func TestBatchUpsertMethods(t *testing.T) {
	type upsertFunc func(ctx context.Context, articles []*Article) error

	tests := []struct {
		name      string
		getFunc   func(repo *ArticleRepository) upsertFunc
		checkTags bool
	}{
		{
			name:      "UpsertArticlesBatch",
			getFunc:   func(repo *ArticleRepository) upsertFunc { return repo.UpsertArticlesBatch },
			checkTags: true,
		},
		{
			name:      "UpsertArticlesUnnestWithoutTags",
			getFunc:   func(repo *ArticleRepository) upsertFunc { return repo.UpsertArticlesUnnestWithoutTags },
			checkTags: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := setupPostgresTestDB(t)
			t.Cleanup(func() {
				truncateArticles(t, db)
				db.Close()
			})
			repo := NewArticleRepository(db)

			existingArticle, _ := NewArticle(
				"uuid-existing-1", "Test Title", "url",
				"content", "author", SourceDevTo,
				[]string{"go", "testing"}, time.Now(),
			)
			insertTestArticle(t, db, existingArticle)

			existingArticle.Title = "Test Title Updated"
			existingArticle.Content = "content updated"
			existingArticle.Tags = []string{"go", "testing", "programming"}
			newArticle, _ := NewArticle(
				"uuid-new-1", "Test Title New", "url",
				"content new", "author", SourceDevTo,
				[]string{"go", "testing"}, time.Now(),
			)

			upsertFn := tt.getFunc(repo)
			err := upsertFn(ctx, []*Article{existingArticle, newArticle})
			assert.Nil(t, err)

			updatedArticle := getTestArticle(t, db, existingArticle.ExternalID)
			assert.Equal(t, existingArticle.UUID, updatedArticle.UUID)
			assert.Equal(t, existingArticle.Title, updatedArticle.Title)
			assert.Equal(t, existingArticle.Author, updatedArticle.Author)
			if tt.checkTags {
				assert.Equal(t, existingArticle.Tags, updatedArticle.Tags)
			}

			insertedArticle := getTestArticle(t, db, newArticle.ExternalID)
			assert.Equal(t, newArticle.UUID, insertedArticle.UUID)
			assert.Equal(t, newArticle.Title, insertedArticle.Title)
			assert.Equal(t, newArticle.Author, insertedArticle.Author)
		})
	}
}
