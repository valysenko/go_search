package fetcher

import (
	"context"
	"fmt"
	"go_search/pkg/redis"
	"time"
)

const lastFetchRun = "last_fetcher_run"

type Storage struct {
	redis *redis.AppRedis
}

func NewStorage(redis *redis.AppRedis) *Storage {
	return &Storage{
		redis: redis,
	}
}

func (s *Storage) GetLastFetchTime(ctx context.Context) (time.Time, error) {
	val, err := s.redis.Client.Get(ctx, lastFetchRun).Result()
	if err != nil {
		return time.Now().Add(-6 * time.Hour), nil
	}

	fetchTime, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse fetch time: %w", err)
	}

	return fetchTime, nil
}

func (s *Storage) SetLastFetchTime(ctx context.Context, fetchTime time.Time) error {
	timeStr := fetchTime.Format(time.RFC3339)
	err := s.redis.Client.Set(ctx, lastFetchRun, timeStr, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set last fetch time in redis: %w", err)
	}
	return nil
}
