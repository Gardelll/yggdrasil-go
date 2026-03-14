//go:build redis

package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisRawCache implements RawCache backed by Redis.
type redisRawCache struct {
	client redis.Cmdable
	prefix string
}

func newRedisRawCache(client redis.Cmdable, prefix string) *redisRawCache {
	return &redisRawCache{client: client, prefix: prefix}
}

func (r *redisRawCache) key(k string) string {
	return r.prefix + k
}

func (r *redisRawCache) GetRaw(key string) ([]byte, bool) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err != nil {
		return nil, false
	}
	return data, true
}

func (r *redisRawCache) SetRaw(key string, data []byte, ttl time.Duration) error {
	ctx := context.Background()
	return r.client.Set(ctx, r.key(key), data, ttl).Err()
}

func (r *redisRawCache) Remove(key string) bool {
	ctx := context.Background()
	result, err := r.client.Del(ctx, r.key(key)).Result()
	return err == nil && result > 0
}
