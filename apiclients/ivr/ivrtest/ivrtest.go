// Package ivrtest provides helpers for testing connections to the IVR API.
package ivrtest

import _ "embed"

var (
	//go:embed twitch_users/not_streaming.json
	TwitchUsersNotStreamingResp string
	//go:embed twitch_users/streaming.json
	TwitchUsersStreamingResp string
	//go:embed twitch_users/banned.json
	TwitchUsersBannedResp string

	//go:embed twitch_users/not_verified_bot.json
	TwitchUsersNotVerifiedBotResp string
	//go:embed twitch_users/verified_bot.json
	TwitchUsersVerifiedBotResp string

	//go:embed mods_and_vips/none.json
	ModsAndVIPsNoneResp string
	//go:embed mods_and_vips/mods_only.json
	ModsAndVIPsModsOnlyResp string
	//go:embed mods_and_vips/mods_and_vips.json
	ModsAndVIPsModsAndVIPsResp string

	//go:embed founders/none_404.json
	FoundersNone404Resp string
	//go:embed founders/none.json
	FoundersNoneResp string
	//go:embed founders/normal.json
	FoundersNormalResp string

	//go:embed sub_age/current_paid_tier3.json
	SubAgeCurrentPaidTier3Resp string
	//go:embed sub_age/current_gift_tier1.json
	SubAgeCurrentGiftTier1Resp string
	//go:embed sub_age/current_prime.json
	SubAgeCurrentPrimeResp string
	//go:embed sub_age/previous_sub.json
	SubAgePreviousSubResp string
	//go:embed sub_age/never_subbed.json
	SubAgeNeverSubbedResp string
	//go:embed sub_age/404_user.json
	SubAge404UserResp string
	//go:embed sub_age/404_channel.json
	SubAge404ChannelResp string
)
