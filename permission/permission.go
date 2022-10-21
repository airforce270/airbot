// Package permission contains permission levels.
package permission

import (
	"math"
	"strconv"
)

// Level represents a permission level.
type Level uint8

// Name returns the human-readable name of the level.
func (l Level) Name() string {
	name, ok := names[l]
	if !ok {
		return strconv.FormatUint(uint64(l), 10)
	}
	return name
}

// IsElevated indicates whether this Level is an elevated one.
func (l Level) IsElevated() bool {
	return l > Normal
}

const (
	// Owner is the owner/host of the bot.
	Owner Level = math.MaxUint8
	// Admin is a channel admin.
	// This may be a specific administrator or, on Twitch, the broadcaster.
	Admin Level = 100
	// Mod is a moderator.
	Mod Level = 60
	// VIP is a channel VIP.
	VIP Level = 50
	// AboveNormal is an above-normal user.
	// This may be a "regular", or on Twitch, a subscriber.
	AboveNormal Level = 30
	// Normal is a normal user.
	Normal Level = 20
	// Unverified is an unverified user.
	// Not all platforms have this.
	Unverified Level = 10
)

// Authorized returns whether a user is authorized at a given level.
func Authorized(user, required Level) bool {
	return user >= required
}

var names = map[Level]string{
	Owner:       "Owner",
	Admin:       "Admin",
	Mod:         "Mod",
	VIP:         "VIP",
	AboveNormal: "Above Normal",
	Normal:      "Normal",
	Unverified:  "Unverified",
}
