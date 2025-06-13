package devto

import (
	"fmt"
	devtoClient "go_search/internal/devto/client"
	"strconv"
)

func ExampleDevTo() {
	page := 1
	perPage := 30
	client := devtoClient.NewDevToClient()
	for {
		request := devtoClient.NewGetArticlesByTagRequest("scala", page, perPage)

		result, _ := client.GetArticlesByTag(request)
		fmt.Println("received count=" + strconv.Itoa(len(result)))

		for _, articleSummary := range result {
			request := devtoClient.NewGetArticlesByIdRequest(articleSummary.ID)
			article, _ := client.GetArticleById(request)
			fmt.Println(article.Title)
		}

		page++
		if len(result) < perPage {
			break
		}
	}
}
