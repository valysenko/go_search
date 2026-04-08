package httpfx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go_search/config"
	"go_search/internal/article"
	"go_search/internal/monitoring"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

func ProvideFiberApp(
	lc fx.Lifecycle,
	cfg *config.HttpAppConfig,
	logger *slog.Logger,
	validator *validator.Validate,
	handler *article.ArticleHandler,
	ms *monitoring.PrometheusMetricsService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    120 * time.Second,
		StructValidator: &structValidator{validate: validator},
	})

	api := app.Group("/api/v1")
	api.Use(monitoring.HttpMonitoringMiddleware(ms))
	articles := api.Group("/article")

	articles.Get("/search", handler.SearchArticle)
	articles.Get("/:uuid", handler.GetArticle)

	promMetrics := app.Group("/metrics")
	passHash := sha256.Sum256([]byte(cfg.MetricsAuth.BasicAuthPassword))
	promMetrics.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			cfg.MetricsAuth.BasicAuthUsername: hex.EncodeToString(passHash[:]),
		},
	}))
	promMetrics.Get("", adaptor.HTTPHandler(promhttp.HandlerFor(ms.Registry(), promhttp.HandlerOpts{})))

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

func ProvideMetricsService(cfg *config.HttpAppConfig) *monitoring.PrometheusMetricsService {
	return monitoring.NewPrometheusMetricsService(cfg.Namespace, "fx_http_server_app", cfg.PodName)
}

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}
