package devto

import (
	"context"
	"errors"
	"go_search/internal/article"
	"go_search/internal/provider"
	"go_search/internal/provider/devto/mocks"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	now := time.Now()
	articlesSince := now.Add(-24 * time.Hour)
	tags := []string{"go", "programming"}
	pr := mocks.NewMockDevToProvider(t)
	runner := NewDevToRunner(pr, nullLogger, tags)

	t.Run("success", func(t *testing.T) {
		pr.On("FetchArticles", ctx, articlesSince, provider.Query{Tags: tags}).Return(nil).Once()

		err := runner.Run(ctx, articlesSince)
		assert.Nil(t, err)
		pr.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("abc")
		pr.On("FetchArticles", ctx, articlesSince, provider.Query{Tags: tags}).Return(expectedErr).Once()

		err := runner.Run(ctx, articlesSince)
		assert.NotNil(t, err)
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
	pr := mocks.NewMockDevToProvider(t)
	runner := NewDevToRunner(pr, nullLogger, tags)

	t.Run("success", func(t *testing.T) {
		articlesChan := make(chan *article.Article, 1)
		errChan := make(chan error, 1)

		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{Tags: tags}, mock.Anything).Return(nil).Once()

		go runner.RunConcurrently(ctx, articlesSince, articlesChan, errChan)

		select {
		case err := <-errChan:
			t.Fatalf("expected no error, but got: %v", err)
		case <-time.After(10 * time.Millisecond):
		}

		pr.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		articlesChan := make(chan *article.Article, 1)
		errChan := make(chan error, 1)

		expectedErr := errors.New("abc")
		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{Tags: tags}, mock.Anything).Return(expectedErr).Once()

		go runner.RunConcurrently(ctx, articlesSince, articlesChan, errChan)

		select {
		case err := <-errChan:
			assert.NotNil(t, err)
			e := &provider.ProviderError{}
			assert.ErrorAs(t, err, &e)
			assert.Equal(t, expectedErr, e.Err)
			assert.Equal(t, provider.DevTo, e.Provider)
		case <-time.After(10 * time.Millisecond):
		}

		pr.AssertExpectations(t)
	})
}
