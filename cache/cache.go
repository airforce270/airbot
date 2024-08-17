// Package cache provides an interface to a cache.
package cache

import (
	"time"
)

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
	// Cache key for the platform that bot restart was requested from.
	KeyRestartRequestedOnPlatform = "restart_requested_on_platform"
	// Cache key for the channel that bot restart was requested from.
	KeyRestartRequestedInChannel = "restart_requested_from_channel"
	// Cache key for the ID of the message that requested the bot restart.
	KeyRestartRequestedByMessageID = "restart_requested_by_message"
)

// GlobalSlowmodeKey returns the global slowmode cache key for a platform.
func GlobalSlowmodeKey(platformName string) string {
	return "global_slowmode_" + platformName
}
