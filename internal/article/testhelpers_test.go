package article

import (
	"bytes"
	"context"
	"go_search/pkg/database"
	"go_search/pkg/es"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/stretchr/testify/assert"
)

// PostgreSQL helpers

func setupPostgresTestDB(t *testing.T) *database.AppDB {
	t.Helper()
	db, err := database.InitDB(&database.DBConfig{
		Host:           "go-search-postgres-test",
		Port:           "5432",
		Username:       "root",
		Password:       "root",
		DbName:         "go_search_db-test",
		MaxConns:       10,
		MinConns:       1,
		ConnectTimeout: 5,
	})
	assert.NoError(t, err)
	db.RunMigrations("../../migrations")

	return db
}

func insertTestArticle(t *testing.T, db *database.AppDB, article *Article) {
	t.Helper()
	_, err := db.Postgresql.Exec(context.Background(), `
          INSERT INTO articles (uuid, external_id, created_at, updated_at, published_at, title, url, content, author, tags, source)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		article.UUID, article.ExternalID, article.CreatedAt, article.UpdateddAt,
		article.PublishedAt, article.Title, article.URL, article.Content,
		article.Author, article.Tags, article.Source,
	)
	if err != nil {
		t.Fatalf("insertTestArticle failed: %v", err)
	}
}

func getTestArticle(t *testing.T, db *database.AppDB, externalID string) *Article {
	t.Helper()
	var art Article
	err := db.Postgresql.QueryRow(context.Background(), "SELECT uuid, title, author, content, tags, published_at, url FROM articles WHERE external_id = $1", externalID).Scan(&art.UUID, &art.Title, &art.Author, &art.Content, &art.Tags, &art.PublishedAt, &art.URL)
	if err != nil {
		return nil
	}

	return &art
}

func truncateArticles(t *testing.T, db *database.AppDB) {
	t.Helper()
	_, err := db.Postgresql.Exec(context.Background(), "TRUNCATE articles CASCADE")
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
}

// Elasticsearch helpers

func setupElasticsearchTestDB(t *testing.T) *es.Client {
	t.Helper()
	ctx := context.Background()

	client, err := es.NewClient(&es.ESConfig{
		Addresses: []string{"http://go-search-elasticsearch-test:9200"},
		Index:     "articles-test",
	})
	if err != nil {
		t.Fatalf("ES client failed: %v", err)
	}

	template, err := os.ReadFile("../../deployments/elasticsearch/index-template.json")
	if err != nil {
		t.Fatalf("failed to read index template file: %v", err)
	}

	putIndexTemplate(t, client, "articles-template", template)
	createIndex(t, client, ctx, "articles-test")

	return client
}

func putIndexTemplate(t *testing.T, client *es.Client, name string, templateJSON []byte) {
	t.Helper()
	_, err := client.GetTypedClient().Indices.PutIndexTemplate(name).Raw(bytes.NewReader(templateJSON)).Do(context.Background())
	if err != nil {
		t.Fatalf("putIndexTemplate failed: %v", err)
	}
}

func createIndex(t *testing.T, client *es.Client, ctx context.Context, name string) {
	t.Helper()
	exists, err := client.GetTypedClient().Indices.Exists(name).Do(ctx)
	if err != nil {
		t.Fatalf("check index existence failed: %v", err)
	}
	if exists {
		return
	}
	_, err = client.GetTypedClient().Indices.Create(name).Do(ctx)
	if err != nil {
		t.Fatalf("createIndex failed: %v", err)
	}
}

func insertTestArticleES(t *testing.T, client *es.Client, article *Article) {
	t.Helper()
	ctx := context.Background()
	index := client.GetIndex()

	_, err := client.GetTypedClient().Index(index).Id(article.UUID).Document(article).Do(ctx)
	if err != nil {
		t.Fatalf("insertTestArticleES failed: %v", err)
	}
	// refresh to make document searchable immediately
	_, err = client.GetTypedClient().Indices.Refresh().Index(index).Do(ctx)
	if err != nil {
		t.Fatalf("refresh index failed: %v", err)
	}
}

func truncateArticlesES(t *testing.T, client *es.Client) {
	t.Helper()
	_, err := client.GetTypedClient().DeleteByQuery(client.GetIndex()).
		Query(&types.Query{MatchAll: &types.MatchAllQuery{}}).Do(context.Background())
	if err != nil {
		t.Fatalf("truncateArticlesES failed: %v", err)
	}
}
