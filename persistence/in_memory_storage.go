package persistence

import (
	"context"
	"sync"
	"time"
)

type InMemoryStorage struct {
	data map[string]int
	mu   sync.Mutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]int),
	}
}

func (m *InMemoryStorage) Get(ctx context.Context, key string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, exists := m.data[key]; exists {
		return val, nil
	}
	return 0, nil
}

func (m *InMemoryStorage) Set(ctx context.Context, key string, value int, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	// Optionally, handle TTL expiration logic here
	return nil
}

func (m *InMemoryStorage) Incr(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, exists := m.data[key]; exists {
		m.data[key] = val + 1
	} else {
		m.data[key] = 1
	}
	return nil
}
