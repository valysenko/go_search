package fetcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"go_search/pkg/redis"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func newMockStorage() (*Storage, redismock.ClientMock) {
	mockClient, mock := redismock.NewClientMock()
	storage := NewStorage(&redis.AppRedis{Client: mockClient})
	return storage, mock
}

func TestGetLastFetchTime(t *testing.T) {
	t.Run("returns default time", func(t *testing.T) {
		storage, mock := newMockStorage()
		mock.ExpectGet(lastFetchRun).RedisNil()

		lastFetchTime, err := storage.GetLastFetchTime(context.Background())

		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now().Add(-6*time.Hour), lastFetchTime, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns stored time", func(t *testing.T) {
		storage, mock := newMockStorage()
		expectedTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
		mock.ExpectGet(lastFetchRun).SetVal(expectedTime.Format(time.RFC3339))

		lastFetchTime, err := storage.GetLastFetchTime(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedTime, lastFetchTime)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on invalid time format", func(t *testing.T) {
		storage, mock := newMockStorage()
		mock.ExpectGet(lastFetchRun).SetVal("invalid-time-format")

		_, err := storage.GetLastFetchTime(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse fetch time")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSetLastFetchTime(t *testing.T) {
	t.Run("sets time successfully", func(t *testing.T) {
		storage, mock := newMockStorage()
		fetchTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
		mock.ExpectSet(lastFetchRun, fetchTime.Format(time.RFC3339), time.Duration(0)).SetVal("OK")

		err := storage.SetLastFetchTime(context.Background(), fetchTime)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on redis failure", func(t *testing.T) {
		storage, mock := newMockStorage()
		fetchTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
		mock.ExpectSet(lastFetchRun, fetchTime.Format(time.RFC3339), time.Duration(0)).SetErr(errors.New("connection refused"))

		err := storage.SetLastFetchTime(context.Background(), fetchTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set last fetch time")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
