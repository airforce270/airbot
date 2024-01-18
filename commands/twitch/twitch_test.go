package twitch_test

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr/ivrtest"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/permission"
)

func TestTwitchCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$br"},
			APIResp:    ivrtest.TwitchUsersBannedResp,
			Want: []*base.Message{
				{
					Text:    "Usage: $banreason <user>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason banneduser",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$br banneduser"},
			APIResp:    ivrtest.TwitchUsersBannedResp,
			Want: []*base.Message{
				{
					Text:    "SeaGrade's ban reason: TOS_INDEFINITE",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason nonbanneduser",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$br nonbanneduser"},
			APIResp:    ivrtest.TwitchUsersNotStreamingResp,
			Want: []*base.Message{
				{
					Text:    "xQc is not banned.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$currentgame",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  twitchtest.GetChannelInformationResp,
			Want: []*base.Message{
				{
					Text:    "user1 is currently playing Science&Technology",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.FoundersNormalResp,
			Want: []*base.Message{
				{
					Text:    "user2's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders hasfounders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.FoundersNormalResp,
			Want: []*base.Message{
				{
					Text:    "hasfounders's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders nofounders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.FoundersNoneResp,
			Want: []*base.Message{
				{
					Text:    "nofounders has no founders",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders nofounders404",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.FoundersNone404Resp,
			Want: []*base.Message{
				{
					Text:    "nofounders404 has no founders",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage macroblank1 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgeCurrentPaidTier3Resp,
			Want: []*base.Message{
				{
					Text:    "Macroblank1 is currently subscribed to xQc with a Tier 3 paid subscription (3 days remaining) and is on a 17 month streak",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$sa macroblank1 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgeCurrentPaidTier3Resp,
			Want: []*base.Message{
				{
					Text:    "Macroblank1 is currently subscribed to xQc with a Tier 3 paid subscription (3 days remaining) and is on a 17 month streak",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$sublength macroblank1 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgeCurrentPaidTier3Resp,
			Want: []*base.Message{
				{
					Text:    "Macroblank1 is currently subscribed to xQc with a Tier 3 paid subscription (3 days remaining) and is on a 17 month streak",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage ellagarten xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgeCurrentGiftTier1Resp,
			Want: []*base.Message{
				{
					Text:    "ellagarten is currently subscribed to xQc with a Tier 1 gifted subscription (14 days remaining) and is on a 4 month streak (total: 17 months)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage airforce2700 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgeCurrentPrimeResp,
			Want: []*base.Message{
				{
					Text:    "airforce2700 is currently subscribed to xQc with a Prime subscription (1 day remaining) and is on a 22 month streak",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage @airforce2700 @elis",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgePreviousSubResp,
			Want: []*base.Message{
				{
					Text:    "airforce2700 is not currently subscribed to elis, but was previously subscribed for 4 months",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage airforce2700 hasanabi",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAgeNeverSubbedResp,
			Want: []*base.Message{
				{
					Text:    "airforce2700 is not subscribed to HasanAbi and has not been previously subscribed",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage airforce2700 channelthatdoesntexist",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAge404ChannelResp,
			Want: []*base.Message{
				{
					Text:    "Channel channelthatdoesntexist was not found",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage userthatdoesntexist xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.SubAge404UserResp,
			Want: []*base.Message{
				{
					Text:    "User userthatdoesntexist was not found",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$logs xqc forsen",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "forsen's logs in xqc's chat: https://logs.ivr.fi/?channel=xqc&username=forsen",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$logs",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Usage: $logs <channel> <user>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.ModsAndVIPsModsAndVIPsResp,
			Want: []*base.Message{
				{
					Text:    "user2's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods otherchannel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.ModsAndVIPsModsAndVIPsResp,
			Want: []*base.Message{
				{
					Text:    "otherchannel's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods nomods",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.ModsAndVIPsNoneResp,
			Want: []*base.Message{
				{
					Text:    "nomods has no mods",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$title",
					UserID:  "user2",
					User:    "user2",
					Channel: "user1",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$title user1"},
			APIResp:    twitchtest.GetChannelInformationResp,
			Want: []*base.Message{
				{
					Text:    "user1's title: TwitchDevMonthlyUpdate//May6,2021",
					Channel: "user1",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbot",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$verifiedbot otherchannel",
				"$vb",
				"$vb otherchannel",
			},
			APIResp: ivrtest.TwitchUsersVerifiedBotResp,
			Want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbot notverified",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$vb notverified"},
			APIResp:    ivrtest.TwitchUsersNotVerifiedBotResp,
			Want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbotquiet",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$verifiedbotquiet otherchannel",
				"$verifiedbotq",
				"$verifiedbotq otherchannel",
				"$vbquiet",
				"$vbquiet otherchannel",
				"$vbq",
				"$vbq otherchannel",
			},
			APIResp: ivrtest.TwitchUsersVerifiedBotResp,
			Want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbotquiet notverified",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$verifiedbotq notverified",
				"$vbquiet notverified",
				"$vbq notverified",
			},
			APIResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			Want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.ModsAndVIPsModsAndVIPsResp,
			Want: []*base.Message{
				{
					Text:    "user2's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips otherchannel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.ModsAndVIPsModsAndVIPsResp,
			Want: []*base.Message{
				{
					Text:    "otherchannel's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips novips",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  ivrtest.ModsAndVIPsNoneResp,
			Want: []*base.Message{
				{
					Text:    "novips has no VIPs",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}
