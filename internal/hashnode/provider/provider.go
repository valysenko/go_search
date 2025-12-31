package provider

import (
	"context"
	"fmt"
	"go_search/pkg/hashnode"
	"time"
)

type Hashnode struct {
	client *hashnode.HashnodeClient
}

func NewHashnode(client *hashnode.HashnodeClient) *Hashnode {
	return &Hashnode{
		client: client,
	}
}

func (hn *Hashnode) Name() string {
	return "Hashnode"
}

func (hn *Hashnode) FetchArticles(ctx context.Context, articlesSince time.Time, tagSlug string) ([]hashnode.FeedPostConnection, error) {
	//slug := "programming"
	sortBy := hashnode.SortByRecent
	first := 10
	var after *string

L:
	for {
		request := hashnode.NewGetArticlesByTagRequest(tagSlug, first, sortBy, after)
		responseData, err := hn.client.GetPostsByTag(request)
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
			if edge.Node.PublishedAt.Before(articlesSince) {
				fmt.Println("END")
				break L
			}
			fmt.Printf("#### %d #####\n", i)
			post := edge.Node
			fmt.Println("Title:", post.Title)
			fmt.Println("Url:", post.URL)
			fmt.Println("Published At:", post.PublishedAt.GoString())
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

	return nil, nil
}
