// Package seventvtest provides helpers for testing connections to the 7TV API.
package seventvtest

import (
	_ "embed"
	"time"

	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/google/go-cmp/cmp"
)

var (
	//go:embed mutate_emote/success.json
	MutateEmoteSuccessResp string
	//go:embed mutate_emote/already_exists.json
	MutateEmoteAlreadyExistsResp string
	//go:embed mutate_emote/id_not_enabled.json
	MutateEmoteIDNotEnabledResp string
	//go:embed mutate_emote/id_not_found.json
	MutateEmoteIDNotFoundResp string
	//go:embed mutate_emote/not_authorized.json
	MutateEmoteNotAuthorizedResp string

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
