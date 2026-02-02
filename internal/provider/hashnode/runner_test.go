package hashnode

import (
	"context"
	"errors"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/internal/provider/hashnode/mocks"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)
	tags := []string{"go", "programming"}

	pr := mocks.NewMockHashnodeProvider(t)
	runner := NewHashnodeRunner(pr, nullLogger, tags, 2)

	t.Run("success", func(t *testing.T) {
		pr.On("FetchArticles", ctx, articlesSince, provider.Query{TagSlug: "go"}).Return(nil).Once()
		pr.On("FetchArticles", ctx, articlesSince, provider.Query{TagSlug: "programming"}).Return(nil).Once()

		err := runner.Run(ctx, articlesSince)
		assert.Nil(t, err)
		pr.AssertExpectations(t)
	})

	t.Run("error on second tag", func(t *testing.T) {
		expectedErr := errors.New("abc")
		pr.On("FetchArticles", ctx, articlesSince, provider.Query{TagSlug: "go"}).Return(nil).Once()
		pr.On("FetchArticles", ctx, articlesSince, provider.Query{TagSlug: "programming"}).Return(expectedErr).Once()

		err := runner.Run(ctx, articlesSince)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		pr.AssertExpectations(t)
	})
}

func TestRunConcurrently(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)
	tags := []string{"go", "programming"}

	pr := mocks.NewMockHashnodeProvider(t)
	runner := NewHashnodeRunner(pr, nullLogger, tags, 2)

	t.Run("success", func(t *testing.T) {
		articlesChan := make(chan<- *article.Article, 2)
		defer close(articlesChan)
		errChan := make(chan error, 3)

		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "go"}, articlesChan).Return(nil).Once()
		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "programming"}, articlesChan).Return(nil).Once()

		done := make(chan struct{})
		go func() {
			runner.RunConcurrently(ctx, articlesSince, articlesChan, errChan)
			close(done)
		}()

		// wait runner to finish
		<-done

		select {
		case err := <-errChan:
			t.Fatalf("expected no error, but got: %v", err)
		default: // should not be errors
		}

		pr.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		expectedErr := errors.New("abc")
		articlesChan := make(chan<- *article.Article, 2)
		defer close(articlesChan)
		errChan := make(chan error, 3)

		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "go"}, articlesChan).Return(nil).Once()
		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "programming"}, articlesChan).Return(expectedErr).Once()

		done := make(chan struct{})
		go func() {
			runner.RunConcurrently(ctx, articlesSince, articlesChan, errChan)
			close(done)
		}()

		// wait runner to finish
		<-done

		select {
		case err := <-errChan:
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, expectedErr)
		default: // should not be errors
		}

		pr.AssertExpectations(t)
	})
}
