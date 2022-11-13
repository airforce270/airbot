// Package cache provides an interface to the local Redis cache.
package cache

import (
	"context"
	"fmt"
	"log"
	"time"

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

const (
	// lastSentTwitchMessageExpiration is the duration the last sent message should remain in the cache.
	// (Twitch blocks messages that are twice in a row in a 30-second period of time)
	lastSentTwitchMessageExpiration = time.Duration(30) * time.Second
	lastSentTwitchMessageKey        = "twitch_last_twitch_message"
)

// FetchLastSentTwitchMessage retrieves the last sent message on Twitch.
func FetchLastSentTwitchMessage(cdb *redis.Client) (string, error) {
	val, err := cdb.Get(context.Background(), lastSentTwitchMessageKey).Result()
	log.Printf("fetched %s", val)
	return val, err
}

// StoreLastSentTwitchMessage stores the last sent message on Twitch.
func StoreLastSentTwitchMessage(cdb *redis.Client, message string) error {
	log.Printf("storing %s", message)
	return cdb.Set(context.Background(), lastSentTwitchMessageKey, message, lastSentTwitchMessageExpiration).Err()
}

func slowmodeKey(p base.Platform) string {
	return fmt.Sprintf("global_slowmode_%s", p.Name())
}
