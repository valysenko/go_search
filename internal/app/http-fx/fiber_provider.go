package httpfx

import (
	"context"
	"go_search/config"
	"go_search/internal/article"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func ProvideFiberApp(
	lc fx.Lifecycle,
	cfg *config.HttpAppConfig,
	logger *slog.Logger,
	validator *validator.Validate,
	handler *article.ArticleHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    120 * time.Second,
		StructValidator: &structValidator{validate: validator},
	})

	api := app.Group("/api/v1")
	articles := api.Group("/article")

	articles.Get("/search", handler.SearchArticle)
	articles.Get("/:uuid", handler.GetArticle)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("starting server", "port", cfg.AppPort)
				if err := app.Listen(":" + cfg.AppPort); err != nil {
					logger.Error("server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down server")
			return app.Shutdown()
		},
	})
	return app
}

func ProvideValidator() *validator.Validate {
	return validator.New()
}

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}
