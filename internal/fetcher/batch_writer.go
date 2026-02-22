package fetcher

import (
	"context"
	"fmt"
	"go_search/internal/article"
	"log/slog"
	"time"
)

type DbBatchWriter struct {
	articleRepository ArticleRepository
	logger            *slog.Logger
	batchSize         int
}

func NewDbBatchWriter(articleRepository ArticleRepository, logger *slog.Logger, batchSize int) *DbBatchWriter {
	return &DbBatchWriter{
		articleRepository: articleRepository,
		logger:            logger,
		batchSize:         batchSize,
	}
}

func (bi *DbBatchWriter) Run(ctx context.Context, articlesChan <-chan *article.Article, errChan chan<- error) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	batch := make([]*article.Article, 0, bi.batchSize)
	flush := func(flushCtx context.Context) {
		if len(batch) == 0 {
			return
		}

		if err := bi.articleRepository.UpsertArticlesBatch(flushCtx, batch); err != nil {
			errChan <- fmt.Errorf("batch insert failed: %w", err)
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
