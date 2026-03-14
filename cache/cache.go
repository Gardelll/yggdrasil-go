package cache

import "time"

// Cache is the basic cache interface used by Session, RegToken, and ProfileKey caches.
type Cache[V any] interface {
	Get(key string) (V, bool)
	Set(key string, value V, ttl time.Duration) error
	Remove(key string) bool
}

// IndexedCache extends Cache with a reverse index, used by TokenService
// to look up tokens by profileId.
type IndexedCache[V any] interface {
	Cache[V]
	SetWithIndex(key string, value V, ttl time.Duration, indexKey string) error
	RemoveByIndex(indexKey string) []string
	GetByIndex(indexKey string) []string
}

// Config holds cache backend configuration.
type Config struct {
	Driver        string `ini:"cache_driver"`
	RedisMode     string `ini:"redis_mode"`     // standalone, sentinel, cluster
	RedisAddress  string `ini:"redis_address"`   // standalone/sentinel: single addr; cluster: comma-separated
	RedisPassword string `ini:"redis_password"`
	RedisDB       int    `ini:"redis_db"`
	RedisPrefix   string `ini:"redis_key_prefix"`

	// Sentinel-specific
	RedisMasterName string `ini:"redis_master_name"`

	// Cluster-specific (addresses are comma-separated in RedisAddress)
}

// RawCache is a type-erased cache interface used internally for provider registration.
type RawCache interface {
	GetRaw(key string) ([]byte, bool)
	SetRaw(key string, data []byte, ttl time.Duration) error
	Remove(key string) bool
}

// RawIndexedCache extends RawCache with index operations.
type RawIndexedCache interface {
	RawCache
	SetRawWithIndex(key string, data []byte, ttl time.Duration, indexKey string) error
	RemoveByIndex(indexKey string) []string
	GetByIndex(indexKey string) []string
}
