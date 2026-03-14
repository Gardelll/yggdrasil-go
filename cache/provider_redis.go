//go:build redis

package cache

import (
	"log"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisOnce      sync.Once
	redisCmdable   redis.Cmdable
)

func getOrCreateRedisClient(cfg *Config) redis.Cmdable {
	redisOnce.Do(func() {
		mode := strings.ToLower(cfg.RedisMode)
		switch mode {
		case "cluster":
			addrs := strings.Split(cfg.RedisAddress, ",")
			for i := range addrs {
				addrs[i] = strings.TrimSpace(addrs[i])
			}
			redisCmdable = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    addrs,
				Password: cfg.RedisPassword,
			})
			log.Printf("Redis Cluster 缓存客户端已初始化: %s", cfg.RedisAddress)
		case "sentinel":
			addrs := strings.Split(cfg.RedisAddress, ",")
			for i := range addrs {
				addrs[i] = strings.TrimSpace(addrs[i])
			}
			redisCmdable = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:    cfg.RedisMasterName,
				SentinelAddrs: addrs,
				Password:      cfg.RedisPassword,
				DB:            cfg.RedisDB,
			})
			log.Printf("Redis Sentinel 缓存客户端已初始化: master=%s, sentinels=%s", cfg.RedisMasterName, cfg.RedisAddress)
		default: // standalone
			redisCmdable = redis.NewClient(&redis.Options{
				Addr:     cfg.RedisAddress,
				Password: cfg.RedisPassword,
				DB:       cfg.RedisDB,
			})
			log.Printf("Redis 缓存客户端已初始化: %s", cfg.RedisAddress)
		}
	})
	return redisCmdable
}

func init() {
	rawCacheFactory = func(cfg *Config, name string) RawCache {
		if cfg.Driver != "redis" {
			return nil
		}
		client := getOrCreateRedisClient(cfg)
		prefix := cfg.RedisPrefix + name + ":"
		return newRedisRawCache(client, prefix)
	}
	rawIndexedCacheFactory = func(cfg *Config, name string) RawIndexedCache {
		if cfg.Driver != "redis" {
			return nil
		}
		client := getOrCreateRedisClient(cfg)
		prefix := cfg.RedisPrefix + name + ":"
		return newRedisRawIndexedCache(client, prefix)
	}
}
