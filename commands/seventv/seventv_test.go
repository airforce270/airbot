package seventv_test

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/seventv/seventvtest"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/permission"
)

func Test7TVCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: nil,
			Want: []*base.Message{
				{
					Text:    "Usage: $7tv add <emote id> [alias]",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: nil,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add fake7tvemoteid somealias",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteSuccessResp,
			},
			Want: []*base.Message{
				{
					Text:    "Added 7TV emote somealias to Twitch/user2",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteSuccessResp,
			},
			Want: []*base.Message{
				{
					Text:    "Added 7TV emote fake7tvemoteid to Twitch/user2",
					Channel: "user2",
				},
			},
		},

		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteNotAuthorizedResp,
			},
			Want: []*base.Message{
				{
					Text:    "Please add me as a 7TV editor if you'd like me to update emotes :)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteAlreadyExistsResp,
			},
			Want: []*base.Message{
				{
					Text:    "Emote is already enabled",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv add fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteIDNotFoundResp,
			},
			Want: []*base.Message{
				{
					Text:    "Emote not found",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv emotecount",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				twitchtest.GetUsersResp,
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
			},
			Want: []*base.Message{
				{
					Text:    "user1 has 3 emotes on 7TV",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv emotecount airforce2700",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				twitchtest.GetUsersResp,
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
			},
			Want: []*base.Message{
				{
					Text:    "airforce2700 has 3 emotes on 7TV",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv remove",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: nil,
			Want: []*base.Message{
				{
					Text:    "Usage: $7tv remove <emote id>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv remove fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: nil,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv remove fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteSuccessResp,
			},
			Want: []*base.Message{
				{
					Text:    "Removed 7TV emote fake7tvemoteid from Twitch/user2",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv remove fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteNotAuthorizedResp,
			},
			Want: []*base.Message{
				{
					Text:    "Please add me as a 7TV editor if you'd like me to update emotes :)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv remove fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteIDNotEnabledResp,
			},
			Want: []*base.Message{
				{
					Text:    "Emote is not enabled",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv remove fake7tvemoteid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform: commandtest.TwitchPlatform,
			APIResps: []string{
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
				seventvtest.MutateEmoteIDNotFoundResp,
			},
			Want: []*base.Message{
				{
					Text:    "Emote not found",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}
