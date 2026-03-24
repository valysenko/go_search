package fetcher

import (
	"context"
	"errors"
	"go_search/internal/fetcher/mocks"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunSequential(t *testing.T) {
	ctx := context.Background()
	articleRepo := &mocks.MockArticleRepository{}
	batchWriter := &mocks.MockBatchWriter{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	fetcherParams := &FetcherParams{
		MaxConcurrentProviders: 1,
		MaxConcurrentDbWriters: 1,
		ArticlesChanBatchSize:  10,
		ErrorsChanBatchSize:    10,
	}

	t.Run("without errors", func(t *testing.T) {
		runner1 := &mocks.MockProviderRunner{}
		runner2 := &mocks.MockProviderRunner{}
		fetcherStorage := &mocks.MockFetcherStorage{}
		fetcher := NewFetcher(articleRepo, fetcherStorage, batchWriter, logger, fetcherParams, runner1, runner2)

		articlesFrom := time.Now().Add(-24 * time.Hour)
		fetcherStorage.On("GetLastFetchTime", ctx).Return(articlesFrom, nil)
		runner1.On("Run", ctx, articlesFrom).Return(nil)
		runner2.On("Run", ctx, articlesFrom).Return(nil)
		fetcherStorage.On("SetLastFetchTime", ctx, mock.Anything).Return(nil)

		res, err := fetcher.RunSequential(ctx)
		assert.Nil(t, err)
		assert.Nil(t, res.Errors)
	})

	t.Run("with errors", func(t *testing.T) {
		runner1 := &mocks.MockProviderRunner{}
		runner2 := &mocks.MockProviderRunner{}
		fetcherStorage := &mocks.MockFetcherStorage{}
		fetcher := NewFetcher(articleRepo, fetcherStorage, batchWriter, logger, fetcherParams, runner1, runner2)

		articlesFrom := time.Now().Add(-24 * time.Hour)
		fetcherStorage.On("GetLastFetchTime", ctx).Return(articlesFrom, nil)
		expectedErr := errors.New("err")
		runner1.On("Run", ctx, articlesFrom).Return(expectedErr)
		runner2.On("Run", ctx, articlesFrom).Return(nil)

		res, err := fetcher.RunSequential(ctx)
		fetcherStorage.AssertNotCalled(t, "SetLastFetchTime")
		assert.Nil(t, err)
		assert.Contains(t, res.Errors, expectedErr)
	})
}

func TestRunConcurrently(t *testing.T) {
	ctx := context.Background()
	articleRepo := &mocks.MockArticleRepository{}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	fetcherParams := &FetcherParams{
		MaxConcurrentProviders: 1,
		MaxConcurrentDbWriters: 1,
		ArticlesChanBatchSize:  10,
		ErrorsChanBatchSize:    10,
	}

	t.Run("without errors", func(t *testing.T) {
		runner1 := &mocks.MockProviderRunner{}
		runner2 := &mocks.MockProviderRunner{}
		fetcherStorage := &mocks.MockFetcherStorage{}
		batchWriter := &mocks.MockBatchWriter{}
		fetcher := NewFetcher(articleRepo, fetcherStorage, batchWriter, logger, fetcherParams, runner1, runner2)

		articlesFrom := time.Now().Add(-24 * time.Hour)
		fetcherStorage.On("GetLastFetchTime", ctx).Return(articlesFrom, nil)
		batchWriter.On("Run", ctx, mock.AnythingOfType("<-chan *article.Article"), mock.AnythingOfType("chan<- error")).Return()
		// runners return immediately which unblocks the "cleanup" part of the fetcher
		runner1.On("RunConcurrently", ctx, articlesFrom, mock.AnythingOfType("chan<- *article.Article"), mock.AnythingOfType("chan<- error")).Return()
		runner2.On("RunConcurrently", ctx, articlesFrom, mock.AnythingOfType("chan<- *article.Article"), mock.AnythingOfType("chan<- error")).Return()
		fetcherStorage.On("SetLastFetchTime", ctx, mock.Anything).Return(nil)

		res, err := fetcher.RunConcurrently(ctx)
		assert.Nil(t, err)
		assert.Nil(t, res.Errors)
	})

	t.Run("with providers error", func(t *testing.T) {
		runner1 := &mocks.MockProviderRunner{}
		runner2 := &mocks.MockProviderRunner{}
		fetcherStorage := &mocks.MockFetcherStorage{}
		batchWriter := &mocks.MockBatchWriter{}
		fetcher := NewFetcher(articleRepo, fetcherStorage, batchWriter, logger, fetcherParams, runner1, runner2)

		articlesFrom := time.Now().Add(-24 * time.Hour)
		expectedWriterErr := errors.New("writer failed")
		expectedProvErr := errors.New("runner failed")

		fetcherStorage.On("GetLastFetchTime", ctx).Return(articlesFrom, nil)
		batchWriter.On("Run", ctx, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				errChan := args.Get(2).(chan<- error)
				errChan <- expectedWriterErr
			}).Return()

		// runner1 sends an error to errChan
		runner1.On("RunConcurrently", ctx, articlesFrom, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				errChan := args.Get(3).(chan<- error)
				errChan <- expectedProvErr
			}).Return()

		// runner2 succeeds
		runner2.On("RunConcurrently", ctx, articlesFrom, mock.Anything, mock.Anything).Return()

		res, err := fetcher.RunConcurrently(ctx)
		assert.Nil(t, err)
		assert.Contains(t, res.Errors, expectedProvErr)
		assert.Contains(t, res.Errors, expectedWriterErr)
	})

	t.Run("with storage error", func(t *testing.T) {
		runner1 := &mocks.MockProviderRunner{}
		runner2 := &mocks.MockProviderRunner{}
		fetcherStorage := &mocks.MockFetcherStorage{}
		batchWriter := &mocks.MockBatchWriter{}
		fetcher := NewFetcher(articleRepo, fetcherStorage, batchWriter, logger, fetcherParams, runner1, runner2)

		expectedErr := errors.New("err")
		fetcherStorage.On("GetLastFetchTime", ctx).Return(time.Time{}, expectedErr)

		_, err := fetcher.RunConcurrently(ctx)

		batchWriter.AssertNotCalled(t, "Run")
		runner1.AssertNotCalled(t, "RunConcurrently")
		runner2.AssertNotCalled(t, "RunConcurrently")
		fetcherStorage.AssertNotCalled(t, "SetLastFetchTime")
		assert.ErrorContains(t, err, expectedErr.Error())
	})
}
