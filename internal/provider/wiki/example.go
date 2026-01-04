package wiki

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	wikiCLientPkg "go_search/pkg/wiki"
	"strconv"
	"time"
)

func RunExampleWithOneQuery() {
	client := wikiCLientPkg.NewWikiClient(10)
	category := "Physics"
	s := "2025-07-25T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	var pages []*wikiCLientPkg.WikiPage
	request := wikiCLientPkg.NewGetCategoryMembersWithGeneratorRequest(category)
L:
	for {
		response, err := client.GetAllCategoryMembersWithPageContent(context.Background(), request)
		if err != nil {
			panic(err)
		}

		for _, page := range response.Query.Pages {
			if page.Revisions[0].Timestamp.Before(articlesFrom) {
				fmt.Println("Reached articles before", page.Revisions[0].Timestamp.GoString())
				break L
			}
			pages = append(pages, &page)
			fmt.Println(page.Title, " - ", page.Revisions[0].Timestamp.GoString())
		}

		// problem: no cmcontinue in response, only rvcontinue
		if response.Continue.CmContinue == "" {
			break L
		}

		request.GcmContinue = response.Continue.CmContinue
	}
}

func RunExampleWithTwoQueries() {
	client := wikiCLientPkg.NewWikiClient(10)
	category := "Physics"
	s := "2025-07-25T00:00:00Z"
	articlesFrom, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	pr := NewWiki(client)
	result, err := pr.FetchArticles(context.Background(), articlesFrom, provider.Query{Category: category})
	i := 0
	for _, article := range result {
		fmt.Println("#" + strconv.Itoa(i) + " - " + article.PublishedAt.GoString() + " - " + article.URL + " - " + article.Author + " - " + article.Title + " - " + fmt.Sprintf("%v", article.Tags))
		i++
	}
}
