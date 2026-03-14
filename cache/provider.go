package cache

import "log"

// rawCacheFactory / rawIndexedCacheFactory are set by build-tag init() (e.g. provider_redis.go)
// to provide alternative cache backends.
var rawCacheFactory func(cfg *Config, name string) RawCache
var rawIndexedCacheFactory func(cfg *Config, name string) RawIndexedCache

// NewCache creates a Cache instance based on configuration.
func NewCache[V any](cfg *Config, name string, capacity int) Cache[V] {
	if cfg != nil && cfg.Driver != "memory" && rawCacheFactory != nil {
		if raw := rawCacheFactory(cfg, name); raw != nil {
			return &typedCache[V]{raw: raw}
		}
		log.Printf("警告: 缓存驱动 %s 不可用, 回退到内存缓存: %s", cfg.Driver, name)
	}
	return NewMemoryCache[V](capacity)
}

// NewIndexedCache creates an IndexedCache instance based on configuration.
func NewIndexedCache[V any](cfg *Config, name string, capacity int) IndexedCache[V] {
	if cfg != nil && cfg.Driver != "memory" && rawIndexedCacheFactory != nil {
		if raw := rawIndexedCacheFactory(cfg, name); raw != nil {
			return &typedIndexedCache[V]{raw: raw}
		}
		log.Printf("警告: 缓存驱动 %s 不可用, 回退到内存缓存: %s", cfg.Driver, name)
	}
	return NewMemoryIndexedCache[V](capacity)
}
