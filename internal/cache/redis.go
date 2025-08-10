package cache

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"task/internal/metrics"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	cli       *redis.Client
	jitterSec int
}

func NewRedisStore(cli *redis.Client, jitterSec int) *RedisStore {
	return &RedisStore{cli: cli, jitterSec: jitterSec}
}

func (s *RedisStore) GetTaskView(key string) (*TaskView, bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	b, err := s.cli.Get(ctx, key).Bytes()
	if err == redis.Nil {
		metrics.CacheGetTotal.WithLabelValues("redis", "miss").Inc()
		return nil, false, "redis", nil
	}
	if err != nil {
		metrics.CacheGetTotal.WithLabelValues("redis", "error").Inc()
		return nil, false, "redis", err
	}
	var tv TaskView
	if err := json.Unmarshal(b, &tv); err != nil {
		metrics.CacheGetTotal.WithLabelValues("redis", "error").Inc()
		return nil, false, "redis", err
	}
	metrics.CacheGetTotal.WithLabelValues("redis", "hit").Inc()
	return &tv, true, "redis", nil
}

func (s *RedisStore) SetTaskView(key string, view *TaskView, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	b, err := json.Marshal(view)
	if err != nil {
		return err
	}
	return s.cli.Set(ctx, key, b, s.withJitter(ttl)).Err()
}

func (s *RedisStore) SetNull(key string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	b, _ := json.Marshal(&TaskView{NotFound: true})
	return s.cli.Set(ctx, key, b, s.withJitter(ttl)).Err()
}

func (s *RedisStore) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	return s.cli.Del(ctx, key).Err()
}

func (s *RedisStore) withJitter(ttl time.Duration) time.Duration {
	if s.jitterSec <= 0 {
		return ttl
	}
	j := time.Duration(rand.Intn(s.jitterSec)) * time.Second
	return ttl + j
}
