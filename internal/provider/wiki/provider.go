package wiki

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/wiki"
	"strconv"
	"time"
)

const AuthorWikiCollaborators = "wiki collaborators"

type ArticleRepository interface {
	UpsertArticle(ctx context.Context, article *article.Article) error
}

type Wiki struct {
	client *wiki.WikiClient
	repo   ArticleRepository
}

func NewWiki(client *wiki.WikiClient, repo ArticleRepository) *Wiki {
	return &Wiki{
		client: client,
		repo:   repo,
	}
}

func (w *Wiki) Name() string {
	return "Wiki"
}

func (d *Wiki) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error {
	cmContinue := "" // retrieved from response
	numArticles := 0
	request := wiki.NewGetCategoryMembersRequest(query.Category, cmContinue)
L:
	for {

		categoryMemberResponse, err := d.client.GetCategoryMembers(ctx, request)
		if err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("fetch cancelled: %w", ctx.Err())
			}
			fmt.Println("API error:", err)
			continue
		}

		for _, item := range categoryMemberResponse.Query.CategoryMembers {
			if item.Timestamp.Before(articlesSince) {
				break L
			}

			pageId := strconv.Itoa(item.PageID)
			page, _ := d.client.GetArticleContent(ctx, wiki.NewGetArticleContentRequest(pageId))

			article, err := article.NewArticle(
				strconv.Itoa(page.GetArticleID(pageId)),
				page.GetArticleTitle(pageId),
				page.GetArticleUrl(pageId),
				page.GetArticleExtract(pageId),
				AuthorWikiCollaborators,
				article.SourceWiki,
				[]string{query.Category},
				item.Timestamp,
			)
			if err != nil {
				continue
			}

			// Temporary solution for testing. Bad performance
			err = d.repo.UpsertArticle(ctx, article)
			if err != nil {
				fmt.Println(err)
				continue
			}

			numArticles++
		}

		if categoryMemberResponse.Continue.CmContinue == "" {
			break L
		}

		request.CmContinue = categoryMemberResponse.Continue.CmContinue
	}

	fmt.Println("fetched " + strconv.Itoa(numArticles) + " wiki articles")
	return nil
}

func (w *Wiki) FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error {
	cmContinue := ""
	numArticles := 0
	request := wiki.NewGetCategoryMembersRequest(query.Category, cmContinue)

L:
	for {
		categoryMemberResponse, err := w.client.GetCategoryMembers(ctx, request)
		if err != nil {
			return err
		}

		for _, item := range categoryMemberResponse.Query.CategoryMembers {
			if item.Timestamp.Before(articlesSince) {
				break L
			}

			pageId := strconv.Itoa(item.PageID)
			page, _ := w.client.GetArticleContent(ctx, wiki.NewGetArticleContentRequest(pageId))

			art, err := article.NewArticle(
				strconv.Itoa(page.GetArticleID(pageId)),
				page.GetArticleTitle(pageId),
				page.GetArticleUrl(pageId),
				page.GetArticleExtract(pageId),
				AuthorWikiCollaborators,
				article.SourceWiki,
				[]string{query.Category},
				item.Timestamp,
			)
			if err != nil {
				continue
			}

			// send to articlesChan OR cancel. goroutine should not be blocked if noone reads from articlesChan and can be finished by ctx.Done()
			select {
			case articlesChan <- art:
				numArticles++
			case <-ctx.Done():
				return fmt.Errorf("cancelled during wiki articles fetching: %w", ctx.Err())
			}
		}

		if categoryMemberResponse.Continue.CmContinue == "" {
			break L
		}

		request.CmContinue = categoryMemberResponse.Continue.CmContinue
	}

	fmt.Println("fetched " + strconv.Itoa(numArticles) + " wiki articles")
	return nil
}
