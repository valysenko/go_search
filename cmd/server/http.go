package main

import (
	"go_search/config"
	"go_search/internal/app/http"
)

func main() {
	appConfig := config.InitHttpAppConfig()

	app := http.NewHttpServerApp(appConfig)
	defer app.Close()
	app.Run()
}
