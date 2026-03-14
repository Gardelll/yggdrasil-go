package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type entry[V any] struct {
	value     V
	expiresAt time.Time
	hasTTL    bool
}

func (e *entry[V]) expired() bool {
	return e.hasTTL && time.Now().After(e.expiresAt)
}

// MemoryCache is an in-memory LRU-based Cache implementation.
type MemoryCache[V any] struct {
	cache *lru.Cache
	mu    sync.RWMutex
}

// NewMemoryCache creates a new in-memory LRU cache with the given capacity.
func NewMemoryCache[V any](capacity int) *MemoryCache[V] {
	c, _ := lru.New(capacity)
	return &MemoryCache[V]{cache: c}
}

func (m *MemoryCache[V]) Get(key string) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.cache.Get(key)
	if !ok {
		var zero V
		return zero, false
	}
	e := v.(*entry[V])
	if e.expired() {
		m.cache.Remove(key)
		var zero V
		return zero, false
	}
	return e.value, true
}

func (m *MemoryCache[V]) Set(key string, value V, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := &entry[V]{value: value}
	if ttl > 0 {
		e.hasTTL = true
		e.expiresAt = time.Now().Add(ttl)
	}
	m.cache.Add(key, e)
	return nil
}

func (m *MemoryCache[V]) Remove(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cache.Remove(key)
}
