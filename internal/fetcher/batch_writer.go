package fetcher

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"log/slog"
	"time"
)

type BatchWriterError struct {
	Msg string
	Err error
}

func (e *BatchWriterError) Error() string {
	return fmt.Sprintf("batch writer error: %s: %s", e.Msg, e.Err)
}

type BatchWriterMetrics interface {
	IncrementRunArticlesTotal(provider, category string, runId string)
}

type DbBatchWriter struct {
	articleRepository ArticleRepository
	logger            *slog.Logger
	batchSize         int
	metrics           BatchWriterMetrics
}

func NewDbBatchWriter(articleRepository ArticleRepository, logger *slog.Logger, metrics BatchWriterMetrics, batchSize int) *DbBatchWriter {
	return &DbBatchWriter{
		articleRepository: articleRepository,
		logger:            logger,
		batchSize:         batchSize,
		metrics:           metrics,
	}
}

func (bi *DbBatchWriter) Run(ctx context.Context, articlesChan <-chan *article.Article, errChan chan<- error, runId string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	batch := make([]*article.Article, 0, bi.batchSize)
	flush := func(flushCtx context.Context) {
		if len(batch) == 0 {
			return
		}

		if err := bi.articleRepository.UpsertArticlesBatch(flushCtx, batch); err != nil {
			errChan <- &BatchWriterError{Msg: "batch insert failed", Err: err}
		} else {
			bi.logger.Info("inserted batch of articles", "batch_size", len(batch))
		}

		batch = make([]*article.Article, 0, bi.batchSize)
	}

	for {
		select {
		case article, ok := <-articlesChan:
			if !ok {
				flush(ctx)
				return
			}
			batch = append(batch, article)
			for _, tag := range article.Tags {
				bi.metrics.IncrementRunArticlesTotal(article.Source.String(), tag, runId)
			}
			if len(batch) >= bi.batchSize {
				flush(ctx)
			}
		case <-ticker.C:
			flush(ctx)
		case <-ctx.Done():
			// context is cancelled. need to pass new context to the writer to finish the current batch insert successfully
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			flush(cleanupCtx)
			cancel()
			return
		}
	}
}
