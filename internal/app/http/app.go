package http

import (
	"context"
	"fmt"
	"go_search/config"
	"go_search/pkg/database"
	"go_search/pkg/es"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type structValidator struct {
	validate *validator.Validate
}

// Validator needs to implement the Validate method
func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

type HttpServerApp struct {
	server     *fiber.App
	db         *database.AppDB
	es         *es.Client
	logger     *slog.Logger
	serverPort string
}

func NewHttpServerApp(cfg *config.HttpAppConfig) *HttpServerApp {
	appDb := database.InitDB(&database.DBConfig{
		Host:           cfg.PostgreSqlConfig.Host,
		Port:           cfg.PostgreSqlConfig.Port,
		Username:       cfg.PostgreSqlConfig.Username,
		Password:       cfg.PostgreSqlConfig.Password,
		DbName:         cfg.PostgreSqlConfig.DbName,
		MaxConns:       cfg.PostgreSqlConfig.MaxConns,
		MinConns:       cfg.PostgreSqlConfig.MinConns,
		ConnectTimeout: cfg.PostgreSqlConfig.ConnectTimeout,
	})
	appDb.RunMigrations("./migrations")

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

	fCfg := fiber.Config{
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    120 * time.Second,
		StructValidator: &structValidator{validate: validator.New()},
	}
	fiber := fiber.New(fCfg)

	app := &HttpServerApp{
		db:         appDb,
		es:         esClient,
		logger:     logger,
		server:     fiber,
		serverPort: cfg.AppPort,
	}

	app.setupRoutes()

	return app
}

func (app *HttpServerApp) setupRoutes() {
	handler := NewArticleHandler(app)
	apiV1 := app.server.Group("/api/v1")
	article := apiV1.Group("/article")

	article.Get("/search", handler.SearchArticle)
	article.Get("/:uuid", handler.GetArticle)
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
