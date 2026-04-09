package fetcherfx

import (
	"context"
	"go_search/internal/fetcher"
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module("fetcher-fx",
	// main fetcher dependency
	fx.Provide(
		fx.Annotate(
			ProvideFetcher,
			fx.ParamTags(``, ``, ``, ``, ``, ``, `group:"runners"`),
		),
	),
	// fetcher parts
	fx.Provide(
		ProvideDevtoProvider,
		ProvideHashnodeProvider,
		ProvideWikiProvider,
		fx.Annotate(
			ProvideDevtoRunner,
			fx.As(new(fetcher.ProviderRunner)),
			fx.ResultTags(`group:"runners"`),
		),
		fx.Annotate(
			ProvideHashnodeRunner,
			fx.As(new(fetcher.ProviderRunner)),
			fx.ResultTags(`group:"runners"`),
		),
		fx.Annotate(
			ProvideWikiRunner,
			fx.As(new(fetcher.ProviderRunner)),
			fx.ResultTags(`group:"runners"`),
		),
	),
	// infra dependencies
	fx.Provide(
		ProvideConfig,
		ProvideLogger,
		ProvideDatabase,
		ProvideArticleRepository,
		ProvideDbBatchWriter,
		ProvideRedis,
		ProvideFetcherStorage,
		ProvideFetcherMetrics,
	),
	// in order to run Fetcher - need to invoke any function which depends on it
	fx.Invoke(RunFetcherJob),
)

func Run() {
	// FX:
	// - builds dependency graph
	// - calls providers to initialize dependencies in correct order
	// - calls invoke functions  (runs fetcher job)
	// - blocks until SIGINT/SIGTERM
	// - runs OnStop hooks (cleanup DB, Redis)
	fx.New(Module).Run()
}

func RunFetcherJob(fa *fetcher.Fetcher, logger *slog.Logger, shutdowner fx.Shutdowner) error {
	ctx := context.Background()
	result, err := fa.RunConcurrently(ctx)

	if err != nil {
		logger.Error("error in running fetcher",
			"error", err,
		)
		return err
	}

	logger.Info("fetch completed", "duration", result.Duration.Seconds())

	if len(result.Errors) > 0 {
		logger.Warn("runner errors occurred during execution",
			"error_count", len(result.Errors),
		)
		for _, runnerErr := range result.Errors {
			logger.Warn("runner error", "err", runnerErr)
		}
	}

	// terminate fx app after fetcher job is done
	return shutdowner.Shutdown()
}
