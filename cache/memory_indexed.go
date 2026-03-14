package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

// MemoryIndexedCache is an in-memory LRU cache with a reverse index.
type MemoryIndexedCache[V any] struct {
	cache *lru.Cache
	index map[string]map[string]struct{} // indexKey -> set of cache keys
	mu    sync.RWMutex
}

// NewMemoryIndexedCache creates a new indexed in-memory LRU cache.
func NewMemoryIndexedCache[V any](capacity int) *MemoryIndexedCache[V] {
	c, _ := lru.New(capacity)
	return &MemoryIndexedCache[V]{
		cache: c,
		index: make(map[string]map[string]struct{}),
	}
}

func (m *MemoryIndexedCache[V]) Get(key string) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.cache.Get(key)
	if !ok {
		var zero V
		return zero, false
	}
	e := v.(*indexedEntry[V])
	if e.expired() {
		m.cache.Remove(key)
		m.removeFromIndex(e.indexKey, key)
		var zero V
		return zero, false
	}
	return e.value, true
}

func (m *MemoryIndexedCache[V]) Set(key string, value V, ttl time.Duration) error {
	return m.SetWithIndex(key, value, ttl, "")
}

func (m *MemoryIndexedCache[V]) SetWithIndex(key string, value V, ttl time.Duration, indexKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If key already exists, remove old index entry
	if old, ok := m.cache.Get(key); ok {
		oldEntry := old.(*indexedEntry[V])
		if oldEntry.indexKey != "" {
			m.removeFromIndex(oldEntry.indexKey, key)
		}
	}

	e := &indexedEntry[V]{value: value, indexKey: indexKey}
	if ttl > 0 {
		e.hasTTL = true
		e.expiresAt = time.Now().Add(ttl)
	}
	m.cache.Add(key, e)

	if indexKey != "" {
		if m.index[indexKey] == nil {
			m.index[indexKey] = make(map[string]struct{})
		}
		m.index[indexKey][key] = struct{}{}
	}
	return nil
}

func (m *MemoryIndexedCache[V]) Remove(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.cache.Get(key); ok {
		e := v.(*indexedEntry[V])
		if e.indexKey != "" {
			m.removeFromIndex(e.indexKey, key)
		}
	}
	return m.cache.Remove(key)
}

func (m *MemoryIndexedCache[V]) RemoveByIndex(indexKey string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys, ok := m.index[indexKey]
	if !ok {
		return nil
	}
	removed := make([]string, 0, len(keys))
	for k := range keys {
		m.cache.Remove(k)
		removed = append(removed, k)
	}
	delete(m.index, indexKey)
	return removed
}

func (m *MemoryIndexedCache[V]) GetByIndex(indexKey string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys, ok := m.index[indexKey]
	if !ok {
		return nil
	}
	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}
	return result
}

// removeFromIndex removes a key from an index entry (must be called with lock held).
func (m *MemoryIndexedCache[V]) removeFromIndex(indexKey, key string) {
	if set, ok := m.index[indexKey]; ok {
		delete(set, key)
		if len(set) == 0 {
			delete(m.index, indexKey)
		}
	}
}

type indexedEntry[V any] struct {
	value     V
	expiresAt time.Time
	hasTTL    bool
	indexKey  string
}

func (e *indexedEntry[V]) expired() bool {
	return e.hasTTL && time.Now().After(e.expiresAt)
}
