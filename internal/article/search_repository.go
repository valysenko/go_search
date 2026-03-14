package article

import (
	"context"
	"encoding/json"
	"go_search/pkg/database"
	"go_search/pkg/es"
	"log/slog"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type ArticleSearchRepository struct {
	esClient *es.Client
	db       *database.AppDB
	logger   *slog.Logger
}

func NewArticleSearchRepository(esClient *es.Client, db *database.AppDB, logger *slog.Logger) *ArticleSearchRepository {
	return &ArticleSearchRepository{
		esClient: esClient,
		db:       db,
		logger:   logger,
	}
}

func (repo *ArticleSearchRepository) SearchArticle(ctx context.Context, query string, limit int) ([]IndexArticle, error) {
	typedClient := repo.esClient.GetTypedClient()
	tokens := strings.Fields(strings.ToLower(query))
	fuzziness := "AUTO"

	res, err := typedClient.Search().
		Index(repo.esClient.GetIndex()).
		SourceIncludes_("uuid", "title", "author", "tags", "published_at").
		Request(&search.Request{
			Size: ptr(limit),
			Query: &types.Query{
				Bool: &types.BoolQuery{
					Should: []types.Query{
						// title field conditions
						{MatchPhrase: map[string]types.MatchPhraseQuery{
							"title": {Query: query, Boost: ptrF32(5.0)},
						}},
						{Match: map[string]types.MatchQuery{
							"title": {Query: query, Fuzziness: &fuzziness, Boost: ptrF32(3.0)},
						}},
						// author field conditions
						{Term: map[string]types.TermQuery{
							"author": {Value: query, Boost: ptrF32(2.0)},
						}},
						// tags field conditions
						{Terms: &types.TermsQuery{
							TermsQuery: map[string]types.TermsQueryField{
								"tags.raw": tokens,
							},
							Boost: ptrF32(4.0),
						}},
						{Match: map[string]types.MatchQuery{
							"tags": {Query: query, Fuzziness: &fuzziness, Boost: ptrF32(2.0)},
						}},
						// content field conditions
						{Match: map[string]types.MatchQuery{
							"content": {Query: query, Fuzziness: &fuzziness, Boost: ptrF32(1.5)},
						}},
					},
					MinimumShouldMatch: "1",
				},
			},
			Highlight: &types.Highlight{
				PreTags:  []string{"[["},
				PostTags: []string{"]]"},
				Fields: map[string]types.HighlightField{
					"content": {
						FragmentSize:      ptr(100),
						NumberOfFragments: ptr(1),
					},
				},
			},
		}).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	r := make([]IndexArticle, 0, len(res.Hits.Hits))

	for _, hit := range res.Hits.Hits {
		art := IndexArticle{}
		err := json.Unmarshal([]byte(hit.Source_), &art)
		if err != nil {
			repo.logger.Error("Error unmarshalling search result", "error", err)
			continue
		}
		if highlights, ok := hit.Highlight["content"]; ok && len(highlights) > 0 {
			art.Highlight = highlights[0]
		}

		r = append(r, art)
	}

	return r, err
}

func (repo *ArticleSearchRepository) GetArticleByUuid(ctx context.Context, uuid string) (*Article, error) {
	var art Article
	err := repo.db.Postgresql.QueryRow(ctx, "SELECT uuid, title, author, content, tags, published_at, url FROM articles WHERE uuid = $1", uuid).Scan(&art.UUID, &art.Title, &art.Author, &art.Content, &art.Tags, &art.PublishedAt, &art.URL)
	if err != nil {
		return nil, err
	}

	return &art, nil
}

func ptr[T any](v T) *T {
	return &v
}

func ptrF32(v float64) *float32 {
	f := float32(v)
	return &f
}
