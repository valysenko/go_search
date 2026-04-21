package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go_search/config"
	"go_search/internal/monitoring"
	"go_search/pkg/database"
	"go_search/pkg/es"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type structValidator struct {
	validate *validator.Validate
}

// Validator needs to implement the Validate method
func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

type HttpServerApp struct {
	server      *fiber.App
	db          *database.AppDB
	es          *es.Client
	logger      *slog.Logger
	ms          *monitoring.PrometheusMetricsService
	serverPort  string
	metricsAuth config.MetricsAuth
}

func NewHttpServerApp(cfg *config.HttpAppConfig) *HttpServerApp {
	appName := "http_server_app"
	db, err := database.InitDB(&database.DBConfig{
		Host:           cfg.PostgreSqlConfig.Host,
		Port:           cfg.PostgreSqlConfig.Port,
		Username:       cfg.PostgreSqlConfig.Username,
		Password:       cfg.PostgreSqlConfig.Password,
		DbName:         cfg.PostgreSqlConfig.DbName,
		MaxConns:       cfg.PostgreSqlConfig.MaxConns,
		MinConns:       cfg.PostgreSqlConfig.MinConns,
		ConnectTimeout: cfg.PostgreSqlConfig.ConnectTimeout,
	})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	db.RunMigrations("./migrations")

	esClient, err := es.NewClient(&es.ESConfig{
		Addresses: []string{
			cfg.ElasticsearchConfig.Host,
		},
		Index: cfg.ElasticsearchConfig.Index,
	})
	if err != nil {
		fmt.Printf("err creating the client: %s\n", err)
		panic(1)
	}
	err = esClient.Ping(context.Background())
	if err != nil {
		fmt.Printf("err pinging the client: %s\n", err)
		panic(1)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	ms := NewMetricsService(cfg.Namespace, appName, cfg.PodName)

	fCfg := fiber.Config{
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    120 * time.Second,
		StructValidator: &structValidator{validate: validator.New()},
	}
	fiber := fiber.New(fCfg)

	app := &HttpServerApp{
		db:          db,
		es:          esClient,
		logger:      logger,
		server:      fiber,
		serverPort:  cfg.AppPort,
		ms:          ms,
		metricsAuth: cfg.MetricsAuth,
	}

	app.setupRoutes()

	return app
}

func (app *HttpServerApp) setupRoutes() {
	handler := NewArticleHandler(app)

	apiV1 := app.server.Group("/api/v1")
	apiV1.Use(monitoring.HttpMonitoringMiddleware(app.ms))
	article := apiV1.Group("/article")

	article.Get("/search", handler.SearchArticle)
	article.Get("/:uuid", handler.GetArticle)

	promMetrics := app.server.Group("/metrics")
	passHash := sha256.Sum256([]byte(app.metricsAuth.BasicAuthPassword))
	promMetrics.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			app.metricsAuth.BasicAuthUsername: hex.EncodeToString(passHash[:]),
		},
	}))
	promMetrics.Get("", adaptor.HTTPHandler(promhttp.HandlerFor(app.ms.Registry(), promhttp.HandlerOpts{})))
}

func (app *HttpServerApp) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		app.logger.Info("starting server")
		if err := app.server.Listen(fmt.Sprintf(":%s", app.serverPort)); err != nil {
			app.logger.Info(fmt.Sprintf("unable to start server - %s", err))
		}

	}()

	<-quit
	app.logger.Info("shutting down server")

	return app.server.Shutdown()
}

func (app *HttpServerApp) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.db.Postgresql.Close()
	app.es.Close(ctx)
}
