package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	c := NewMemoryCache[string](100)
	testCacheBasic(t, c, "memory")
}

func TestMemoryCacheTTL(t *testing.T) {
	c := NewMemoryCache[string](100)
	testCacheTTL(t, c, "memory")
}

func TestMemoryIndexedCache(t *testing.T) {
	c := NewMemoryIndexedCache[string](100)
	testIndexedCacheBasic(t, c, "memory_indexed")
}

func TestNewCacheDefault(t *testing.T) {
	// nil config should return memory cache
	c := NewCache[string](nil, "test", 100)
	if c == nil {
		t.Fatal("NewCache returned nil")
	}
	_ = c.Set("k", "v", 0)
	v, ok := c.Get("k")
	if !ok || v != "v" {
		t.Errorf("Expected 'v', got '%s' (ok=%v)", v, ok)
	}
}

func TestNewIndexedCacheDefault(t *testing.T) {
	c := NewIndexedCache[string](nil, "test", 100)
	if c == nil {
		t.Fatal("NewIndexedCache returned nil")
	}
	_ = c.SetWithIndex("k1", "v1", 0, "idx1")
	keys := c.GetByIndex("idx1")
	if len(keys) != 1 || keys[0] != "k1" {
		t.Errorf("Expected [k1], got %v", keys)
	}
}

func testCacheBasic(t *testing.T, c Cache[string], label string) {
	t.Helper()
	// Set and Get
	_ = c.Set("key1", "value1", 0)
	v, ok := c.Get("key1")
	if !ok || v != "value1" {
		t.Errorf("[%s] Get: expected 'value1', got '%s' (ok=%v)", label, v, ok)
	}
	// Get missing
	_, ok = c.Get("missing")
	if ok {
		t.Errorf("[%s] expected miss for 'missing'", label)
	}
	// Remove
	c.Remove("key1")
	_, ok = c.Get("key1")
	if ok {
		t.Errorf("[%s] expected miss after remove", label)
	}
}

func testCacheTTL(t *testing.T, c Cache[string], label string) {
	t.Helper()
	_ = c.Set("ttl_key", "ttl_val", 50*time.Millisecond)
	v, ok := c.Get("ttl_key")
	if !ok || v != "ttl_val" {
		t.Errorf("[%s] TTL: expected value before expiry", label)
	}
	time.Sleep(100 * time.Millisecond)
	_, ok = c.Get("ttl_key")
	if ok {
		t.Errorf("[%s] TTL: expected miss after expiry", label)
	}
}

func testIndexedCacheBasic(t *testing.T, c IndexedCache[string], label string) {
	t.Helper()
	// SetWithIndex
	_ = c.SetWithIndex("a1", "v1", 0, "group1")
	_ = c.SetWithIndex("a2", "v2", 0, "group1")
	_ = c.SetWithIndex("b1", "v3", 0, "group2")

	// GetByIndex
	keys := c.GetByIndex("group1")
	if len(keys) != 2 {
		t.Errorf("[%s] expected 2 keys in group1, got %d", label, len(keys))
	}

	// RemoveByIndex
	removed := c.RemoveByIndex("group1")
	if len(removed) != 2 {
		t.Errorf("[%s] expected 2 removed, got %d", label, len(removed))
	}
	_, ok := c.Get("a1")
	if ok {
		t.Errorf("[%s] expected miss for a1 after RemoveByIndex", label)
	}
	// group2 unaffected
	v, ok := c.Get("b1")
	if !ok || v != "v3" {
		t.Errorf("[%s] expected group2 key to survive", label)
	}
}
