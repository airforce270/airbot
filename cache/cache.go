// Package cache provides an interface to the local Redis cache.
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/airforce270/airbot/base"

	"github.com/redis/go-redis/v9"
)

// Instance is an instance of the cache.
var Instance Cache = nil

// A Cache stores and retrieves simple key-value data quickly.
type Cache interface {
	// StoreBool stores a bool value with no expiration.
	StoreBool(key string, value bool) error
	// StoreExpiringBool stores a bool value with an expiration.
	// If the key does not exist, false will be returned.
	StoreExpiringBool(key string, value bool, expiration time.Duration) error
	// FetchBool fetches a bool value.
	// If the key does not exist, false will be returned.
	FetchBool(key string) (bool, error)

	// StoreString stores a string value with no expiration.
	StoreString(key, value string) error
	// StoreExpiringString stores a string value with an expiration.
	// If the key does not exist, an empty string will be returned.
	StoreExpiringString(key, value string, expiration time.Duration) error
	// FetchString fetches a string value.
	// If the key does not exist, an empty string will be returned.
	FetchString(key string) (string, error)
}

const (
	// Cache key for the last sent Twitch message.
	KeyLastSentTwitchMessage = "twitch_last_sent_message"
)

// GlobalSlowmodeKey returns the global slowmode cache key for a platform.
func GlobalSlowmodeKey(p base.Platform) string {
	return fmt.Sprintf("global_slowmode_%s", p.Name())
}

// RedisCache implements Cache for a real Redis database.
type RedisCache struct {
	r *redis.Client
}

func (c *RedisCache) StoreBool(key string, value bool) error {
	return c.StoreExpiringBool(key, value, 0)
}
func (c *RedisCache) StoreExpiringBool(key string, value bool, expiration time.Duration) error {
	return c.r.Set(context.Background(), key, value, expiration).Err()
}
func (c *RedisCache) FetchBool(key string) (bool, error) {
	resp, err := c.r.Get(context.Background(), key).Bool()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return resp, nil
}
func (c *RedisCache) StoreString(key, value string) error {
	return c.StoreExpiringString(key, value, 0)
}
func (c *RedisCache) StoreExpiringString(key, value string, expiration time.Duration) error {
	return c.r.Set(context.Background(), key, value, expiration).Err()
}
func (c *RedisCache) FetchString(key string) (string, error) {
	val, err := c.r.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

// NewRedisCache creates a new Redis-backed Cache.
func NewRedisCache() RedisCache {
	return RedisCache{r: redis.NewClient(&redis.Options{Addr: "cache:6379"})}
}
