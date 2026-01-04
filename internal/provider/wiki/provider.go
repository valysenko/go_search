package wiki

import (
	"context"
	"fmt"
	"go_search/internal/provider"
	"go_search/pkg/wiki"
	"strconv"
	"time"
)

const AuthorWikiCollaborators = "wiki collaborators"

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

func (d *Wiki) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) ([]provider.Article, error) {
	response := []provider.Article{}
	cmContinue := "" // retrieved from response
	request := wiki.NewGetCategoryMembersRequest(query.Category, cmContinue)
L:
	for {

		categoryMemberResponse, err := d.client.GetCategoryMembers(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, item := range categoryMemberResponse.Query.CategoryMembers {
			if item.Timestamp.Before(articlesSince) {
				fmt.Println("Reached articles before", item.Timestamp.GoString())
				break L
			}

			pageId := strconv.Itoa(item.PageID)
			page, _ := d.client.GetArticleContent(ctx, wiki.NewGetArticleContentRequest(pageId))
			response = append(response, provider.Article{
				ID:          strconv.Itoa(page.GetArticleID(pageId)),
				Title:       page.GetArticleTitle(pageId),
				URL:         page.GetArticleUrl(pageId),
				Content:     page.GetArticleExtract(pageId),
				PublishedAt: item.Timestamp,
				Author:      AuthorWikiCollaborators,
				Tags:        []string{query.TagSlug},
				Source:      provider.SourceWiki,
			})
		}

		if categoryMemberResponse.Continue.CmContinue == "" {
			break L
		}

		request.CmContinue = categoryMemberResponse.Continue.CmContinue
	}

	return response, nil
}
