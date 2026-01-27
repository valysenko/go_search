package devto

import (
	"context"
	"errors"
	"fmt"
	"go_search/helpers"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/pkg/devto"
	"go_search/pkg/httpclient"
	"log/slog"
	"strconv"

	"time"
)

type ArticleRepository interface {
	UpsertArticle(ctx context.Context, article *article.Article) error
	UpsertArticlesBatch(ctx context.Context, articles []*article.Article) error
	UpsertArticlesUnnestWithoutTags(ctx context.Context, articles []*article.Article) error
}

type Client interface {
	GetLatestArticles(ctx context.Context, request *devto.GetLatestArticlesRequest) ([]devto.ArticleSummary, error)
	GetArticleById(ctx context.Context, request *devto.GetArticlesByIdRequest) (*devto.Article, error)
}

type DevTo struct {
	client Client
	repo   ArticleRepository
	logger *slog.Logger
}

func NewDevToProvider(client Client, repo ArticleRepository, logger *slog.Logger) *DevTo {
	return &DevTo{
		client: client,
		repo:   repo,
		logger: logger,
	}
}

func (d *DevTo) Name() string {
	return "DevTo"
}

func (d *DevTo) FetchArticles(ctx context.Context, articlesSince time.Time, query provider.Query) error {
	page := 1
	perPage := 30
	numArticles := 0
L:
	for {
		request := devto.NewGetLatestArticlesRequest(page, perPage)
		result, err := d.client.GetLatestArticles(ctx, request)

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("devto: cancelled at page %d: %w", page, err)
			}

			// 4xx is allowed
			var clientErr *httpclient.RequestError
			if errors.As(err, &clientErr) && clientErr.Type == httpclient.ErrorTypeServer {
				return fmt.Errorf("devto: server error at page %d: %w", page, err)
			}

			d.logger.Warn("page failed, continuing",
				"page", page,
				"error", err,
			)

			page++
			continue
		}

		i := 0

		articles := make([]*article.Article, 0, 30)

		for _, articleSummary := range result {
			if articlesSince.After(articleSummary.PublishedAt) {
				break L
			}

			if helpers.HasAny(articleSummary.TagList, query.Tags) {
				request := devto.NewGetArticlesByIdRequest(articleSummary.ID)
				sourceArticle, err := d.client.GetArticleById(ctx, request)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return fmt.Errorf("devto: cancelled at page %d: %w", page, err)
					}
					d.logger.Warn("failed to get article",
						"ID", request.ID,
						"error", err,
					)
					continue
				}

				article, err := article.NewArticle(
					strconv.Itoa(sourceArticle.ID),
					sourceArticle.Title,
					sourceArticle.Url,
					sourceArticle.BodyMarkdown,
					sourceArticle.User.Name,
					article.SourceDevTo,
					sourceArticle.TagList,
					sourceArticle.PublishedAt,
				)
				if err != nil {
					d.logger.Warn("failed to create article",
						"ID", request.ID,
						"error", err,
					)
					continue
				}

				articles = append(articles, article)

				i++
				numArticles++
			}
		}

		if len(articles) > 0 {
			err = d.repo.UpsertArticlesUnnestWithoutTags(ctx, articles)
		}

		page++
		if len(result) < perPage {
			break
		}
	}

	return nil
}

func (d *DevTo) FetchArticlesAsync(ctx context.Context, articlesSince time.Time, query provider.Query, articlesChan chan<- *article.Article) error {
	page := 1
	perPage := 30
	numArticles := 0

L:
	for {
		request := devto.NewGetLatestArticlesRequest(page, perPage)
		result, err := d.client.GetLatestArticles(ctx, request)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("devto: cancelled at page %d: %w", page, err)
			}

			// 4xx is allowed
			var clientErr *httpclient.RequestError
			if errors.As(err, &clientErr) && clientErr.Type == httpclient.ErrorTypeServer {
				return fmt.Errorf("devto: server error at page %d: %w", page, err)
			}

			d.logger.Warn("page failed, continuing",
				"page", page,
				"error", err,
			)
			page++
			continue
		}

		for _, articleSummary := range result {
			if articlesSince.After(articleSummary.PublishedAt) {
				break L
			}

			if helpers.HasAny(articleSummary.TagList, query.Tags) {
				request := devto.NewGetArticlesByIdRequest(articleSummary.ID)
				sourceArticle, err := d.client.GetArticleById(ctx, request)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return fmt.Errorf("devto: cancelled at page %d: %w", page, err)
					}
					d.logger.Warn("failed to get article",
						"ID", request.ID,
						"error", err,
					)
					continue
				}

				art, err := article.NewArticle(
					strconv.Itoa(sourceArticle.ID),
					sourceArticle.Title,
					sourceArticle.Url,
					sourceArticle.BodyMarkdown,
					sourceArticle.User.Name,
					article.SourceDevTo,
					sourceArticle.TagList,
					sourceArticle.PublishedAt,
				)
				if err != nil {
					d.logger.Warn("failed to create article",
						"ID", request.ID,
						"error", err,
					)
					continue
				}

				// send to articlesChan OR cancel. goroutine should not be blocked if noone reads from articlesChan and can be finished by ctx.Done()
				select {
				case articlesChan <- art:
					numArticles++
				case <-ctx.Done():
					return fmt.Errorf("cancelled during devto articles fetching: %w", ctx.Err())
				}
			}
		}

		page++
		if len(result) < perPage {
			break
		}
	}

	d.logger.Info("fetched articles",
		"count", numArticles,
	)
	return nil
}
