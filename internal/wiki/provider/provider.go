package provider

import (
	"context"
	"fmt"
	"go_search/pkg/wiki"
	"strconv"
	"time"
)

type Wiki struct {
	client *wiki.WikiClient
}

func NewWiki(client *wiki.WikiClient) *Wiki {
	return &Wiki{
		client: client,
	}
}

func (w *Wiki) Name() string {
	return "Wiki"
}

func (d *Wiki) FetchArticles(ctx context.Context, articlesSince time.Time, category string) error {
	cmContinue := "" // retrieved from response
	request := wiki.NewGetCategoryMembersRequest(category, cmContinue)
L:
	for {

		categoryMemberResponse, err := d.client.GetCategoryMembers(ctx, request)
		if err != nil {
			panic(err)
		}

		for _, item := range categoryMemberResponse.Query.CategoryMembers {
			if item.Timestamp.Before(articlesSince) {
				fmt.Println("Reached articles before", item.Timestamp.GoString())
				break L
			}

			article, _ := d.client.GetArticleContent(ctx, wiki.NewGetArticleContentRequest(strconv.Itoa(item.PageID)))
			fmt.Println(article.Query.Pages[strconv.Itoa(item.PageID)].Title, " - ", item.Timestamp.GoString())
		}

		if categoryMemberResponse.Continue.CmContinue == "" {
			break L
		}

		request.CmContinue = categoryMemberResponse.Continue.CmContinue
	}

	return nil
}
