package httpfx

import (
	"go_search/internal/article"
	"log/slog"
)

func ProvideArticleHandler(
	repo *article.ArticleSearchRepository,
	logger *slog.Logger,
) *article.ArticleHandler {
	return article.NewArticleHandler(repo, logger)
}
