package article

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetArticleHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	repo := &ArticleSearchRepository{
		esClient: nil,
		db:       setupPostgresTestDB(t),
		logger:   logger,
	}
	ah := NewArticleHandler(repo, logger)
	t.Cleanup(func() {
		truncateArticles(t, repo.db)
		repo.db.Close()
	})

	app := fiber.New()
	app.Get("/article/:uuid", ah.GetArticle)

	t.Run("article exists", func(t *testing.T) {
		existingArticle, _ := NewArticle(
			"uuid-existing-1", "Test Title", "url",
			"content", "author", SourceDevTo,
			[]string{"go", "testing"}, time.Now(),
		)
		insertTestArticle(t, repo.db, existingArticle)

		req, err := http.NewRequest("GET", "/article/"+existingArticle.UUID, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, _ := app.Test(req)
		var responseBody Article
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatal("failed to decode response body:", err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, existingArticle.UUID, responseBody.UUID)
		assert.Equal(t, existingArticle.Title, responseBody.Title)
		assert.Equal(t, existingArticle.Content, responseBody.Content)
	})

	t.Run("article does not exist", func(t *testing.T) {
		uuidv7, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodGet, "/article/"+uuidv7.String(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("article bad request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/article/asdasd", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestSearchArticleHandler(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	es := setupElasticsearchTestDB(t)
	t.Cleanup(func() {
		truncateArticlesES(t, es)
		es.Close(ctx)
	})

	repo := &ArticleSearchRepository{
		esClient: es,
		db:       nil,
		logger:   logger,
	}
	ah := NewArticleHandler(repo, logger)

	app := fiber.New()
	app.Get("/article/search", ah.SearchArticle)

	t.Run("article exists", func(t *testing.T) {
		testArticle := &Article{
			UUID:    "uuid-php",
			Title:   "Php language",
			Author:  "John",
			Content: "Php language is a popular programming language",
			Tags:    []string{"php", "programming"},
		}
		insertTestArticleES(t, es, testArticle)

		testArticle = &Article{
			UUID:    "uuid-php-callable",
			Title:   "Callables in PHP",
			Author:  "John",
			Content: "There are callables in PHP and they are very useful",
			Tags:    []string{"php", "callables", "programming"},
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

		params := url.Values{}
		params.Set("query", "callables php")
		req, err := http.NewRequest("GET", "/article/search?"+params.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := app.Test(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseBody []IndexArticle
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatal("failed to decode response body:", err)
		}

		assert.Len(t, responseBody, 2)
		assert.Equal(t, "Callables in PHP", responseBody[0].Title)
		assert.Equal(t, "Php language", responseBody[1].Title)
	})

	t.Run("article does not exist", func(t *testing.T) {
		params := url.Values{}
		params.Set("query", "python")
		req, err := http.NewRequest("GET", "/article/search?"+params.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := app.Test(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
