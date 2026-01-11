package main

import (
	"go_search/internal/article"
	"go_search/internal/provider/devto"
	"go_search/pkg/database"
)

func main() {
	dbConfig := database.NewDBConfig("go-search-postgres", "5432", "root", "root", "go_search_db", 10, 10, 5)
	appDb := database.InitDB(dbConfig)
	appDb.RunMigrations("./migrations")
	defer appDb.Close()

	articleRepository := article.NewArticleRepository(appDb)

	devto.ExampleProvider(appDb, articleRepository)
	// wiki.RunExampleWithTwoQueries(articleRepository)
	// hashnode.ExampleHashnodeProvider(appDb, articleRepository)

	// ctx, _ := context.WithCancel(context.Background())
	// pool := appDb.Postgresql

	// for i := 0; i < 5; i++ {
	// 	go func(count int) {
	// 		_, err := pool.Exec(ctx, ";")
	// 		if err != nil {
	// 			fmt.Fprintf(os.Stderr, "Ping failed: %v\n", err)
	// 			os.Exit(1)
	// 		}
	// 		fmt.Println(count, "Query OK!")
	// 	}(i)
	// }
	// select {}
}
