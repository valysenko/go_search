package http

import (
	"go_search/internal/article"
	"go_search/internal/monitoring"
)

func NewArticleSearchRepository(app *HttpServerApp) *article.ArticleSearchRepository {
	return article.NewArticleSearchRepository(app.es, app.db, app.logger)
}

func NewArticleHandler(app *HttpServerApp) *article.ArticleHandler {
	repo := NewArticleSearchRepository(app)
	return article.NewArticleHandler(repo, app.logger)
}

func NewMetricsService(namespace, subsystem, podName string) *monitoring.PrometheusMetricsService {
	return monitoring.NewPrometheusMetricsService(namespace, subsystem, podName)
}
