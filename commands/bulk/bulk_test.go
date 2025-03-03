package bulk_test

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/pastebin/pastebintest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
)

func TestBulkCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay https://pastebin.com/raw/B7TBjQEy",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen"),
				},
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  pastebintest.MultiLineFetchPasteResp,
			Want: []*base.Message{
				{
					Text:    "line1",
					Channel: "user2",
				},
				{
					Text:    "line2",
					Channel: "user2",
				},
				{
					Text:    "line3",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay https://pastebin.com/raw/B7TBjQEy",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  pastebintest.MultiLineFetchPasteResp,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen"),
				},
			},
			Platform: commandtest.TwitchPlatform,
			APIResp:  pastebintest.MultiLineFetchPasteResp,
			Want: []*base.Message{
				{
					Text:    "Usage: $filesay <pastebin raw URL>",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}
