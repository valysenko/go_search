package fetcher

import (
	"context"
	"go_search/internal/article"
	"go_search/internal/fetcher/mocks"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tests := []struct {
		name        string
		checkCancel bool
		checkClose  bool
	}{
		{
			name:        "Run with cancel",
			checkCancel: true,
		},
		{
			name:       "Run with channel close",
			checkClose: true,
		},
	}

	for _, tt := range tests {
		articleRepo := mocks.NewMockArticleRepository(t)
		metrics := mocks.NewMockBatchWriterMetrics(t)
		batchWriter := NewDbBatchWriter(articleRepo, nullLogger, metrics, 3)

		t.Run(tt.name, func(t *testing.T) {
			var runId string = "100"
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			articlesChan := make(chan *article.Article)
			errChan := make(chan error)

			art1 := &article.Article{UUID: "1", Title: "Article 1", Tags: []string{"tag1"}, Source: article.SourceDevTo}
			art2 := &article.Article{UUID: "2", Title: "Article 2", Tags: []string{"tag2", "tag5"}, Source: article.SourceHashnode}
			art3 := &article.Article{UUID: "3", Title: "Article 3", Tags: []string{"tag3"}, Source: article.SourceWiki}
			art4 := &article.Article{UUID: "4", Title: "Article 4", Tags: []string{"tag4"}, Source: article.SourceHashnode}

			articleRepo.On("UpsertArticlesBatch", mock.Anything, []*article.Article{art1, art2, art3}).Return(nil)
			articleRepo.On("UpsertArticlesBatch", mock.Anything, []*article.Article{art4}).Return(nil)
			metrics.On("IncrementRunArticlesTotal", "devto", "tag1", runId).Return()
			metrics.On("IncrementRunArticlesTotal", "hashnode", "tag2", runId).Return()
			metrics.On("IncrementRunArticlesTotal", "hashnode", "tag5", runId).Return()
			metrics.On("IncrementRunArticlesTotal", "wiki", "tag3", runId).Return()
			metrics.On("IncrementRunArticlesTotal", "hashnode", "tag4", runId).Return()

			testChan := make(chan struct{})
			// run the batch writer in a separate goroutine
			go func() {
				batchWriter.Run(ctx, articlesChan, errChan, runId)
				close(testChan)
			}()

			articlesChan <- art1
			articlesChan <- art2
			articlesChan <- art3
			articlesChan <- art4

			if tt.checkCancel {
				cancel()
			} else if tt.checkClose {
				close(articlesChan)
			}

			<-testChan

			articleRepo.AssertExpectations(t)
			metrics.AssertExpectations(t)
		})
	}
}
