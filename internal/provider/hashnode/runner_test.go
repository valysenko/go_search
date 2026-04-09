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
	"github.com/stretchr/testify/mock"
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

type runConcurrentlyFunc func(ctx context.Context, articlesFrom time.Time, articlesChan chan<- *article.Article, errChan chan<- error)

func TestRunConcurrently(t *testing.T) {
	runner, pr := setupRunner(t)
	checkConcurrentRunner(t, pr, runner.RunConcurrently)
}

func TestRunConcurrentlyWP(t *testing.T) {
	runner, pr := setupRunner(t)
	checkConcurrentRunner(t, pr, runner.RunConcurrentlyWP)
}

func TestRunConcurrentlyWPStreaming(t *testing.T) {
	runner, pr := setupRunner(t)
	checkConcurrentRunner(t, pr, runner.RunConcurrentlyWPStreaming)
}

func setupRunner(t *testing.T) (*HashnodeRunner, *mocks.MockHashnodeProvider) {
	nullLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	tags := []string{"go", "programming"}
	pr := mocks.NewMockHashnodeProvider(t)
	runner := NewHashnodeRunner(pr, nullLogger, tags, 2)
	return runner, pr
}

func checkConcurrentRunner(t *testing.T, pr *mocks.MockHashnodeProvider, runFn runConcurrentlyFunc) {
	ctx := context.Background()
	articlesSince := time.Now().Add(-24 * time.Hour)

	t.Run("success", func(t *testing.T) {
		articlesChan := make(chan *article.Article, 2) // Bidirectional, not send-only
		errChan := make(chan error, 3)

		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "go"}, mock.Anything).Return(nil).Once()
		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "programming"}, mock.Anything).Return(nil).Once()

		runFn(ctx, articlesSince, articlesChan, errChan)

		select {
		case err := <-errChan:
			t.Fatalf("expected no error, got: %v", err)
		default:
		}

		pr.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		expectedErr := errors.New("abc")
		articlesChan := make(chan *article.Article, 2)
		errChan := make(chan error, 3)

		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "go"}, mock.Anything).Return(nil).Once()
		pr.On("FetchArticlesAsync", ctx, articlesSince, provider.Query{TagSlug: "programming"}, mock.Anything).Return(expectedErr).Once()

		runFn(ctx, articlesSince, articlesChan, errChan)

		select {
		case err := <-errChan:
			e := &provider.ProviderError{}
			assert.ErrorAs(t, err, &e)
			assert.Equal(t, expectedErr, e.Err)
			assert.Equal(t, provider.Hashnode, e.Provider)
		default:
			t.Fatal("expected error, got none")
		}

		pr.AssertExpectations(t)
	})
}
