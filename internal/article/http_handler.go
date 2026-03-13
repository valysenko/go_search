package article

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

type SearchQuery struct {
	Query string `query:"query" validate:"required,min=3,max=100"`
	Limit int    `query:"limit" validate:"min=1,max=20"`
}

type ArticleHandler struct {
	searchRepo *ArticleSearchRepository
	logger     *slog.Logger
}

func NewArticleHandler(repo *ArticleSearchRepository, logger *slog.Logger) *ArticleHandler {
	return &ArticleHandler{searchRepo: repo, logger: logger}
}

func (ah *ArticleHandler) GetArticle(c fiber.Ctx) error {
	ctx := context.Background()
	article, err := ah.searchRepo.GetArticleByUuid(ctx, c.Params("uuid"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("not found")
	}

	return c.Status(fiber.StatusOK).JSON(article)
}

func (ah *ArticleHandler) SearchArticle(c fiber.Ctx) error {
	ctx := context.Background()
	req := &SearchQuery{Limit: 10}
	if err := c.Bind().Query(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	searchResult, err := ah.searchRepo.SearchArticle(ctx, req.Query, req.Limit)
	if err != nil {
		ah.logger.Error("Error searching articles", "error", err)
		return c.Status(fiber.StatusInternalServerError).SendString("internal server error")
	}
	if len(searchResult) == 0 {
		return c.Status(fiber.StatusNotFound).SendString("no articles found")
	}

	return c.Status(fiber.StatusOK).JSON(searchResult)
}
