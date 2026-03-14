//go:build redis

package cache

import (
	"testing"
	"time"
)

func getRedisConfig() *Config {
	return &Config{
		Driver:        "redis",
		RedisAddress:  "127.0.0.1:6379",
		RedisPassword: "",
		RedisDB:       0,
		RedisPrefix:   "ygg_test:",
	}
}

func TestRedisCache(t *testing.T) {
	cfg := getRedisConfig()
	c := NewCache[string](cfg, "test_basic", 100)
	// Clean up
	defer c.Remove("key1")
	testCacheBasic(t, c, "redis")
}

func TestRedisCacheTTL(t *testing.T) {
	cfg := getRedisConfig()
	c := NewCache[string](cfg, "test_ttl", 100)
	testCacheTTL(t, c, "redis")
}

func TestRedisIndexedCache(t *testing.T) {
	cfg := getRedisConfig()
	c := NewIndexedCache[string](cfg, "test_idx", 100)
	// Clean up
	defer func() {
		c.RemoveByIndex("group1")
		c.RemoveByIndex("group2")
	}()
	testIndexedCacheBasic(t, c, "redis_indexed")
}

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRedisCacheStruct(t *testing.T) {
	cfg := getRedisConfig()
	c := NewCache[*testStruct](cfg, "test_struct", 100)
	defer c.Remove("s1")

	orig := &testStruct{Name: "test", Value: 42}
	_ = c.Set("s1", orig, 0)

	got, ok := c.Get("s1")
	if !ok {
		t.Fatal("expected hit")
	}
	if got.Name != "test" || got.Value != 42 {
		t.Errorf("expected {test, 42}, got {%s, %d}", got.Name, got.Value)
	}
}

func TestRedisCacheTTLExpiry(t *testing.T) {
	cfg := getRedisConfig()
	c := NewCache[string](cfg, "test_ttl_exp", 100)

	_ = c.Set("exp_key", "exp_val", 100*time.Millisecond)
	v, ok := c.Get("exp_key")
	if !ok || v != "exp_val" {
		t.Errorf("expected value before expiry")
	}
	time.Sleep(200 * time.Millisecond)
	_, ok = c.Get("exp_key")
	if ok {
		t.Errorf("expected miss after TTL expiry")
	}
}
