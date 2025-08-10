package cache

import (
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LocalStore struct {
	lru *expirable.LRU[string, *TaskView]
}

func NewLocalStore(capacity int, defaultTTL time.Duration) *LocalStore {
	return &LocalStore{lru: expirable.NewLRU[string, *TaskView](capacity, nil, defaultTTL)}
}

func (s *LocalStore) GetTaskView(key string) (*TaskView, bool, error) {
	v, ok := s.lru.Get(key)
	return v, ok, nil
}

func (s *LocalStore) SetTaskView(key string, view *TaskView, ttl time.Duration) error {
	s.lru.Add(key, view)
	return nil
}

func (s *LocalStore) SetNull(key string, ttl time.Duration) error {
	s.lru.Add(key, &TaskView{NotFound: true})
	return nil
}

func (s *LocalStore) Delete(key string) error {
	s.lru.Remove(key)
	return nil
}
