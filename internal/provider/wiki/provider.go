package wiki

import (
	"context"
	"errors"
	"fmt"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/httpclient"
	"go_search/pkg/wiki"
	"log/slog"
	"strconv"
	"time"
)

const AuthorWikiCollaborators = "wiki collaborators"

type articleHandler func(ctx context.Context, art *article.Article) error

type ArticleRepository interface {
	UpsertArticle(ctx context.Context, article *article.Article) error
}

type WikiClient interface {
	GetCategoryMembers(ctx context.Context, request *wiki.GetCategoryMembersRequest) (*wiki.CategoryMembersResponse, error)
	GetArticleContent(ctx context.Context, request *wiki.GetArticleContentRequest) (*wiki.ArticleResponse, error)
}

type Wiki struct {
	client WikiClient
	repo   ArticleRepository
	logger *slog.Logger
}

func NewWiki(client WikiClient, logger *slog.Logger, repo ArticleRepository) *Wiki {
	return &Wiki{
		client: client,
		repo:   repo,
		logger: logger,
	}
}

func (w *Wiki) Name() string {
	return "Wiki"
}

func (w *Wiki) fetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query, handler articleHandler) error {
	cmContinue := ""
	numArticles := 0
	request := wiki.NewGetCategoryMembersRequest(query.Category, cmContinue)

L:
	for {
		categoryMemberResponse, err := w.client.GetCategoryMembers(ctx, request)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("wiki: fetch cancelled for category '%s': %w", query.Category, err)
			}
			return fmt.Errorf("wiki: failed to get category members for %q: %w", query.Category, err)
		}

		for _, item := range categoryMemberResponse.Query.CategoryMembers {
			if item.Timestamp.Before(articlesSince) {
				break L
			}

			pageId := strconv.Itoa(item.PageID)
			page, err := w.client.GetArticleContent(ctx, wiki.NewGetArticleContentRequest(pageId))
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return fmt.Errorf("wiki: article fetch cancelled for category '%s': %w", query.Category, err)
				}

				var reqErr *httpclient.RequestError
				if errors.As(err, &reqErr) && reqErr.Type == httpclient.ErrorTypeNetwork {
					return fmt.Errorf("wiki: network error fetching content for %s: %w", pageId, err)
				}

				w.logger.Warn("wiki: skipping page", "page_id", pageId, "error", err)

				continue
			}

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
				w.logger.Warn("wiki: failed to create article from page",
					"page_id", pageId,
					"error", err)
				continue
			}

			if err := handler(ctx, art); err != nil {
				return err
			}

			numArticles++
		}

		if categoryMemberResponse.Continue.CmContinue == "" {
			break L
		}

		request.CmContinue = categoryMemberResponse.Continue.CmContinue
	}

	w.logger.Info("fetched articles",
		"count", numArticles,
		"category", query.Category)

	return nil
}

func (w *Wiki) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error {
	handler := func(ctx context.Context, art *article.Article) error {
		if err := w.repo.UpsertArticle(ctx, art); err != nil {
			w.logger.Warn("wiki: failed to upsert article",
				"article_id", art.ExternalID,
				"error", err)
		}
		return nil
	}

	return w.fetchArticles(ctx, articlesSince, query, handler)
}

func (w *Wiki) FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error {
	handler := func(ctx context.Context, art *article.Article) error {
		select {
		case articlesChan <- art:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("cancelled while sending article %s: %w", art.ExternalID, ctx.Err())
		}
	}

	return w.fetchArticles(ctx, articlesSince, query, handler)
}
