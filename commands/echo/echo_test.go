package echo_test

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
)

func TestEchoCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$commands",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Commands available here: https://github.com/airforce270/airbot/blob/main/docs/commands.md",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$gn",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "FeelsOkayMan <3 gn user1",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$spam 3 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "yo",
					Channel: "user2",
				},
				{
					Text:    "yo",
					Channel: "user2",
				},
				{
					Text:    "yo",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$spam 3 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 5 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 1000 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Max pyramid width is 25",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 5 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$trihard",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$TriHard"},
			Want: []*base.Message{
				{
					Text:    "TriHard 7",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$tuck",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Usage: $tuck <user>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$tuck someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting("forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Bedge user1 tucks someone into bed.",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}
