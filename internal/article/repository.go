package article

import (
	"context"
	"fmt"
	"go_search/pkg/database"
	"time"

	"github.com/jackc/pgx/v4"
)

type ArticleRepository struct {
	db *database.AppDB
}

func NewArticleRepository(db *database.AppDB) *ArticleRepository {
	return &ArticleRepository{
		db: db,
	}
}

func (repo *ArticleRepository) UpsertArticle(ctx context.Context, article *Article) error {
	query := `
		INSERT INTO articles (
			uuid, external_id, created_at, updated_at, published_at, title, url, content, author, tags, source
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (external_id, source) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			content = EXCLUDED.content,
			author = EXCLUDED.author,
			published_at = EXCLUDED.published_at,
			tags = EXCLUDED.tags,
			updated_at = NOW()
		RETURNING uuid
	`

	err := repo.db.Postgresql.QueryRow(ctx, query,
		article.UUID,
		article.ExternalID,
		article.CreatedAt,
		article.UpdateddAt,
		article.PublishedAt,
		article.Title,
		article.URL,
		article.Content,
		article.Author,
		article.Tags,
		article.Source,
	).Scan(&article.UUID)

	if err != nil {
		return fmt.Errorf("failed to upsert article: %w", err)
	}

	return nil
}

// Batching reduces network latency by "pipelining" the commands - it sends a stream of queries all at once (Request + Request + Request -> Response + Response + Response).
// 1 -  queue up multiple SQL statements and their arguments. 2 - send them to the server in one go, and then read the results.
func (repo *ArticleRepository) UpsertArticlesBatch(ctx context.Context, articles []*Article) error {
	if len(articles) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO articles (
			uuid, external_id, created_at, updated_at, published_at, title, url, content, author, tags, source
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (external_id, source) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			content = EXCLUDED.content,
			author = EXCLUDED.author,
			published_at = EXCLUDED.published_at,
			tags = EXCLUDED.tags,
			updated_at = NOW()`

	for _, a := range articles {
		batch.Queue(query,
			a.UUID, a.ExternalID, a.CreatedAt, a.UpdateddAt, a.PublishedAt,
			a.Title, a.URL, a.Content, a.Author, a.Tags, a.Source,
		)
	}

	br := repo.db.Postgresql.SendBatch(ctx, batch)

	defer br.Close()

	for i := 0; i < len(articles); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("error in batch at index %d: %w", i, err)
		}
	}

	return nil
}

// It is expected that it is faster than pgx.Batch because the database engine processes it as a single execution plan rather than a series of individual statements.
// Data is passed as a set of arrays (array of UUIDs, an array of Titles, ...). The UNNEST function takes these arrays and turns them into a temporary table-like structure that you can then INSERT from.
// (!) Problem with multidimensional arrays: unnest cannot infer multidimensional arrays ([][]string for tags) - https://github.com/launchbadge/sqlx/issues/1945
func (repo *ArticleRepository) UpsertArticlesUnnestWithoutTags(ctx context.Context, articles []*Article) error {
	if len(articles) == 0 {
		return nil
	}

	uuids := make([]string, len(articles))
	extIDs := make([]string, len(articles))
	titles := make([]string, len(articles))
	urls := make([]string, len(articles))
	contents := make([]string, len(articles))
	authors := make([]string, len(articles))
	sources := make([]int16, len(articles))
	publishedAts := make([]time.Time, len(articles))
	createdAts := make([]time.Time, len(articles))
	updatedAts := make([]time.Time, len(articles))
	// tags := make([][]string, len(articles))
	// tags := make([]pgtype.TextArray, len(articles))

	for i, a := range articles {
		uuids[i] = a.UUID
		extIDs[i] = a.ExternalID
		titles[i] = a.Title
		urls[i] = a.URL
		contents[i] = a.Content
		authors[i] = a.Author
		sources[i] = int16(a.Source)
		publishedAts[i] = a.PublishedAt
		createdAts[i] = a.CreatedAt
		updatedAts[i] = a.UpdateddAt
		// tags[i] = a.Tags // - pgx cannot infer multidimensional arrays

		// var ta pgtype.TextArray
		// _ = ta.Set(a.Tags) // []string → text[]
		// tags[i] = ta
	}

	query := `
		INSERT INTO articles (
			uuid, external_id, created_at, updated_at, published_at, title, url, content, author, source
		)
		SELECT u.uuid, u.ext_id, u.created_at, u.updated_at, u.published_at, u.title, u.url, u.content, u.author, u.source FROM UNNEST(
			$1::uuid[], $2::varchar(65)[], $3::timestamptz[], $4::timestamptz[], $5::timestamptz[], $6::text[], $7::text[], $8::text[], $9::varchar(255)[], $10::smallint[]
		) AS u(uuid, ext_id, created_at, updated_at, published_at, title, url, content, author, source)
		ON CONFLICT (external_id, source) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			content = EXCLUDED.content,
			author = EXCLUDED.author,
			published_at = EXCLUDED.published_at,
			updated_at = NOW()`

	_, err := repo.db.Postgresql.Exec(ctx, query,
		uuids, extIDs, createdAts, updatedAts, publishedAts, titles, urls, contents, authors, sources,
	)

	if err != nil {
		return fmt.Errorf("failed bulk upsert: %w", err)
	}

	return nil
}
