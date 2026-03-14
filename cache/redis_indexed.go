//go:build redis

package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisRawIndexedCache implements RawIndexedCache backed by Redis.
type redisRawIndexedCache struct {
	client    redis.Cmdable
	prefix    string
	idxPrefix string
}

func newRedisRawIndexedCache(client redis.Cmdable, prefix string) *redisRawIndexedCache {
	return &redisRawIndexedCache{
		client:    client,
		prefix:    prefix,
		idxPrefix: prefix + "idx:",
	}
}

func (r *redisRawIndexedCache) key(k string) string {
	return r.prefix + k
}

func (r *redisRawIndexedCache) idxKey(indexKey string) string {
	return r.idxPrefix + indexKey
}

func (r *redisRawIndexedCache) GetRaw(key string) ([]byte, bool) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err != nil {
		return nil, false
	}
	return data, true
}

func (r *redisRawIndexedCache) SetRaw(key string, data []byte, ttl time.Duration) error {
	return r.SetRawWithIndex(key, data, ttl, "")
}

func (r *redisRawIndexedCache) SetRawWithIndex(key string, data []byte, ttl time.Duration, indexKey string) error {
	ctx := context.Background()
	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.key(key), data, ttl)
	if indexKey != "" {
		pipe.SAdd(ctx, r.idxKey(indexKey), key)
		// Do not set TTL on the index set itself — individual data keys have their own TTL.
		// The index set is cleaned up via RemoveByIndex or when all members are expired.
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *redisRawIndexedCache) Remove(key string) bool {
	ctx := context.Background()
	result, err := r.client.Del(ctx, r.key(key)).Result()
	return err == nil && result > 0
}

func (r *redisRawIndexedCache) RemoveByIndex(indexKey string) []string {
	ctx := context.Background()
	members, err := r.client.SMembers(ctx, r.idxKey(indexKey)).Result()
	if err != nil || len(members) == 0 {
		return nil
	}
	keys := make([]string, len(members))
	for i, m := range members {
		keys[i] = r.key(m)
	}
	pipe := r.client.Pipeline()
	pipe.Del(ctx, keys...)
	pipe.Del(ctx, r.idxKey(indexKey))
	_, _ = pipe.Exec(ctx)
	return members
}

func (r *redisRawIndexedCache) GetByIndex(indexKey string) []string {
	ctx := context.Background()
	members, err := r.client.SMembers(ctx, r.idxKey(indexKey)).Result()
	if err != nil {
		return nil
	}
	return members
}
