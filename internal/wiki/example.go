package wiki

import (
	"fmt"
	wikiCLientPkg "go_search/internal/wiki/client"
	"strconv"
)

func RunExampleWithOneQuery() {
	client := wikiCLientPkg.NewWikiClient()
	category := "Techno"

	categoryMembers, err := client.GetCategoryMembersWithPageContent(wikiCLientPkg.NewGetCategoryMembersWithGeneratorRequest(category))
	if err != nil {
		panic(err)
	}
	for _, member := range categoryMembers {
		fmt.Println(strconv.Itoa(member.PageID) + " - " + member.Title)
	}
}

func RunExampleWithTwoQueries() {
	client := wikiCLientPkg.NewWikiClient()
	category := "Physics"

	categoryMembers, err := client.GetCategoryMembers(wikiCLientPkg.NewGetCategoryMembersRequest(category, ""))
	if err != nil {
		panic(err)
	}

	// for _, member := range categoryMembers {
	// 	fmt.Println(strconv.Itoa(member.PageID) + " - " + member.Title)
	// }

	for _, member := range categoryMembers {
		response, _ := client.GetArticleContent(wikiCLientPkg.NewGetArticleContentRequest(strconv.Itoa(member.PageID)))
		fmt.Println(response.Query.Pages)
	}
}
