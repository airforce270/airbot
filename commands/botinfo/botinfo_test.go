package botinfo_test

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
					Text:    "$bot",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$botinfo",
				"$info",
				"$about",
			},
			Want: []*base.Message{
				{
					Text:    "Beep boop, this is Airbot running as fake-username in user2 with prefix $ on Twitch. Made by airforce2700, source available on GitHub ( $source )",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "For help with a command, use $help <command>. To see available commands, use $commands",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "[ $join ] Tells the bot to join your chat.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help duel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "[ $duel ] Duels another chatter. They have 30 seconds to accept or decline. User-specific cooldown: 5s",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help pyramid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "[ $pyramid ] Makes a pyramid in chat. Max width 25. Channel-wide cooldown: 30s",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "??prefix",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    ";prefix",
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
				"does this bot thingy have one of them prefixes",
				"what is a prefix",
				"forsen prefix",
				"Successfully joined channel iP0G with prefix $",
			},
			Want: nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$source",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Source code for Airbot available at https://github.com/airforce270/airbot",
					Channel: "user2",
				},
			},
		},

		// stats is currently untested due to reliance on low-level syscalls
	}

	commandtest.Run(t, tests)
}
