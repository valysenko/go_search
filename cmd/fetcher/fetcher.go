package main

import (
	"context"
	"go_search/config"
	"go_search/internal/app/fetcher"
)

func main() {
	ctx := context.Background()
	cfg := config.InitConfig()
	app := fetcher.NewFetcherApp(cfg)
	defer app.Close(ctx)
	app.Run(ctx)
}
