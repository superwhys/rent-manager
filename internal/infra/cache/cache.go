package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miebyte/goutils/redisutils"
)

const globalPrefix = "rent-manager"

type optionFn func(*CacheKey)

func WithExpire(duration time.Duration) optionFn {
	return func(ck *CacheKey) {
		ck.expire = duration
	}
}

type CacheKey struct {
	key    string
	expire time.Duration
}

func (c *CacheKey) ExpireTime() time.Duration {
	return c.expire
}

func (c *CacheKey) String() string {
	return c.key
}

func (c *CacheKey) Key() string {
	return c.key
}

func (c *CacheKey) Get(ctx context.Context, client *redisutils.RedisClient) (string, error) {
	return client.Get(ctx, c.key).Result()
}

func (c *CacheKey) Set(ctx context.Context, client *redisutils.RedisClient, value any) error {
	if err := client.Set(ctx, c.key, value, c.expire).Err(); err != nil {
		return fmt.Errorf("set cache(%s) failed: %w", c.key, err)
	}
	return nil
}

func (c *CacheKey) Del(ctx context.Context, client *redisutils.RedisClient) error {
	if err := client.Del(ctx, c.key).Err(); err != nil {
		return fmt.Errorf("del cache(%s) failed: %w", c.key, err)
	}
	return nil
}

func (c *CacheKey) Incr(ctx context.Context, client *redisutils.RedisClient) error {
	if err := client.Incr(ctx, c.key).Err(); err != nil {
		return fmt.Errorf("incr cache(%s) failed: %w", c.key, err)
	}
	return nil
}

type genCacheWithKeyFn func(key string, opts ...optionFn) *CacheKey

func genCacheWithKey(prefix string, expire time.Duration) genCacheWithKeyFn {
	return func(key string, opts ...optionFn) *CacheKey {
		keyItems := []string{globalPrefix, prefix}
		if key != "" {
			keyItems = append(keyItems, key)
		}

		ck := &CacheKey{
			key:    strings.Join(keyItems, ":"),
			expire: expire,
		}
		for _, opt := range opts {
			opt(ck)
		}
		return ck
	}
}

type genCacheFn func(...optionFn) *CacheKey

func genCache(key string, expire time.Duration) genCacheFn {
	return func(opts ...optionFn) *CacheKey {
		ck := &CacheKey{
			key:    key,
			expire: expire,
		}

		for _, opt := range opts {
			opt(ck)
		}

		return ck
	}
}
