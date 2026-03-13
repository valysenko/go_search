package http

import (
	"go_search/internal/article"
)

func NewArticleSearchRepository(app *HttpServerApp) *article.ArticleSearchRepository {
	return article.NewArticleSearchRepository(app.es, app.db, app.logger)
}

func NewArticleHandler(app *HttpServerApp) *article.ArticleHandler {
	repo := NewArticleSearchRepository(app)
	return article.NewArticleHandler(repo, app.logger)
}
