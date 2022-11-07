// Package cache provides an interface to the local Redis cache.
package cache

import (
	"context"
	"fmt"

	"github.com/airforce270/airbot/base"
	"github.com/go-redis/redis/v9"
)

// Instance is an instance of the Redis client.
var Instance *redis.Client = nil

// NewClient creates a new Redis client.
func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "cache:6379"})
}

// FetchSlowmode returns whether a platform is following a global 1-second slowmode.
func FetchSlowmode(p base.Platform, cdb *redis.Client) (bool, error) {
	slowmodeResp, err := cdb.Get(context.Background(), slowmodeKey(p)).Bool()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return slowmodeResp, nil
}

// FetchSlowmode sets whether a platform should follow a global 1-second slowmode.
func SetSlowmode(p base.Platform, cdb *redis.Client, value bool) error {
	return cdb.Set(context.Background(), slowmodeKey(p), value, 0).Err()
}

func slowmodeKey(p base.Platform) string {
	return fmt.Sprintf("global_slowmode_%s", p.Name())
}
