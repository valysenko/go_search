package httpfx

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

var Module = fx.Module("http-server",
	// all dependencies should be provided
	fx.Provide(
		ProvideFiberApp,
		ProvideValidator,
	),
	fx.Provide(
		ProvideArticleHandler,
	),
	fx.Provide(
		ProvideConfig,
		ProvideLogger,
		ProvideDatabase,
		ProvideElasticsearch,
		ProvideArticleSearchRepository,
	),
	// in order to run Fiber Http server - need to invoke any function which depends on it
	fx.Invoke(func(*fiber.App) {}),
)

func Run() {
	// FX:
	// - builds dependency graph
	// - calls providers to initialize dependencies in correct order
	// - calls invoke functions (triggers *fiber.App creation)
	// - runs OnStart hooks (app.Listen starts the server)
	// - blocks until SIGINT/SIGTERM
	// - runs OnStop hooks (app.Shutdown, es.Close)
	fx.New(Module).Run()
}
