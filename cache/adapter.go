package cache

import (
	"encoding/json"
	"time"
)

// typedCache wraps a RawCache with JSON serialization to implement Cache[V].
type typedCache[V any] struct {
	raw RawCache
}

func (c *typedCache[V]) Get(key string) (V, bool) {
	var zero V
	data, ok := c.raw.GetRaw(key)
	if !ok {
		return zero, false
	}
	var v V
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, false
	}
	return v, true
}

func (c *typedCache[V]) Set(key string, value V, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.raw.SetRaw(key, data, ttl)
}

func (c *typedCache[V]) Remove(key string) bool {
	return c.raw.Remove(key)
}

// typedIndexedCache wraps a RawIndexedCache with JSON serialization to implement IndexedCache[V].
type typedIndexedCache[V any] struct {
	raw RawIndexedCache
}

func (c *typedIndexedCache[V]) Get(key string) (V, bool) {
	var zero V
	data, ok := c.raw.GetRaw(key)
	if !ok {
		return zero, false
	}
	var v V
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, false
	}
	return v, true
}

func (c *typedIndexedCache[V]) Set(key string, value V, ttl time.Duration) error {
	return c.SetWithIndex(key, value, ttl, "")
}

func (c *typedIndexedCache[V]) SetWithIndex(key string, value V, ttl time.Duration, indexKey string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.raw.SetRawWithIndex(key, data, ttl, indexKey)
}

func (c *typedIndexedCache[V]) Remove(key string) bool {
	return c.raw.Remove(key)
}

func (c *typedIndexedCache[V]) RemoveByIndex(indexKey string) []string {
	return c.raw.RemoveByIndex(indexKey)
}

func (c *typedIndexedCache[V]) GetByIndex(indexKey string) []string {
	return c.raw.GetByIndex(indexKey)
}
