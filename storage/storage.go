package storage

import (
	"context"
	"sync"
	"time"
)

type Storage interface {
	Allow(ctx context.Context, key string, limit int, duration time.Duration) (bool, error)
}

// In-memory store implementation for testing
type InMemoryStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{tokens: make(map[string]time.Time)}
}

func (s *InMemoryStore) Allow(ctx context.Context, key string, limit int, duration time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Now().After(s.tokens[key]) {
		s.tokens[key] = time.Now().Add(duration)
		return true, nil
	}
	return false, nil
}
