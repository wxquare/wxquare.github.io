package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// LocalCache 本地缓存（L1）
type LocalCache struct {
	data sync.Map
}

type localCacheItem struct {
	value      interface{}
	expiration time.Time
}

func NewLocalCache() *LocalCache {
	cache := &LocalCache{}
	go cache.cleanupExpired()
	return cache
}

func (c *LocalCache) Get(key string) (interface{}, bool) {
	value, ok := c.data.Load(key)
	if !ok {
		return nil, false
	}

	item := value.(localCacheItem)
	if time.Now().After(item.expiration) {
		c.data.Delete(key)
		return nil, false
	}

	return item.value, true
}

func (c *LocalCache) Set(key string, value interface{}, ttl time.Duration) {
	item := localCacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	c.data.Store(key, item)
}

func (c *LocalCache) Delete(key string) {
	c.data.Delete(key)
}

func (c *LocalCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.data.Range(func(key, value interface{}) bool {
			item := value.(localCacheItem)
			if now.After(item.expiration) {
				c.data.Delete(key)
			}
			return true
		})
	}
}

// RedisCache Redis缓存（L2）- 简化实现
type RedisCache struct {
	data sync.Map
}

func NewRedisCache() *RedisCache {
	return &RedisCache{}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, ok := c.data.Load(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	item := value.(localCacheItem)
	if time.Now().After(item.expiration) {
		c.data.Delete(key)
		return nil, fmt.Errorf("key expired: %s", key)
	}

	return json.Marshal(item.value)
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var data interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return err
	}

	item := localCacheItem{
		value:      data,
		expiration: time.Now().Add(ttl),
	}
	c.data.Store(key, item)
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	c.data.Delete(key)
	return nil
}
