// Package seventvtest provides helpers for testing connections to the 7TV API.
package seventvtest

import (
	_ "embed"
	"time"

	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/google/go-cmp/cmp"
)

var (
	//go:embed fetch_uc_by_twitch_uid/small_non_sub.json
	FetchUserConnectionByTwitchUserIdSmallNonSubResp string
	//go:embed fetch_uc_by_twitch_uid/small_sub.json
	FetchUserConnectionByTwitchUserIdSmallSubResp string
	//go:embed fetch_uc_by_twitch_uid/large_sub.json
	FetchUserConnectionByTwitchUserIdLargeSubResp string

	// Transformer is a cmp.Option that transforms a UnixTimeMs to a time.Time.
	Transformer = cmp.Transformer("NativeUnixTimeMs", func(in seventv.UnixTimeMs) time.Time {
		return time.Time(in)
	})
)
