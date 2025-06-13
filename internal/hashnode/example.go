package hashnode

import (
	"fmt"
	hnClient "go_search/internal/hashnode/client"
)

func ExampleHashnode() {
	client := hnClient.NewHashnodeClient()

	slug := "car"
	sortBy := hnClient.SortByRecent
	first := 10
	var after *string

	for {
		request := hnClient.NewGetArticlesByTagRequest(slug, first, sortBy, after)
		responseData, err := client.GetPostsByTag(request)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nTag: %s (%s)\n", responseData.Tag.Name, responseData.Tag.Slug)
		fmt.Printf("Count: %d\n", responseData.Tag.PostsCount)
		fmt.Printf("Has Next Page: %v\n", responseData.Tag.Posts.PageInfo.HasNextPage)
		fmt.Printf("EndCursor: %s\n", responseData.Tag.Posts.PageInfo.EndCursor)
		i := 1
		fmt.Println("Posts:\n")
		for _, edge := range responseData.Tag.Posts.Edges {
			fmt.Printf("#### %d #####\n", i)
			post := edge.Node
			fmt.Println("Title:", post.Title)
			fmt.Println("Url:", post.URL)
			fmt.Println("Published At:", post.PublishedAt)
			fmt.Println("Author Username:", post.Author.Username)
			// fmt.Println("Cursor: %s", edge.Cursor)
			//fmt.Println("Markdown Content:", post.Content.Markdown)
			//fmt.Println("HTML Content:", post.Content.HTML)
			//fmt.Println("Text Content:", post.Content.Text)
			fmt.Printf("#### %d #####\n\n", i)
			i++
		}

		if responseData.Tag.Posts.PageInfo.HasNextPage == false {
			break
		}
		after = &responseData.Tag.Posts.PageInfo.EndCursor
		fmt.Println("")
	}
}
