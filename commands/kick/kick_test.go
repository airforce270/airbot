package kick_test

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/kick/kicktest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/permission"
)

func TestKickCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kickislive brucedropemoff",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$kislive brucedropemoff"},
			APIResp:    kicktest.LargeLiveGetChannelResp,
			Want: []*base.Message{
				{
					Text:    "brucedropemoff is currently live on Kick, streaming Just Chatting to 15274 viewers.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kickislive xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$kislive xqc"},
			APIResp:    kicktest.LargeOfflineGetChannelResp,
			Want: []*base.Message{
				{
					Text:    "xqc is not currently live on Kick.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kicktitle brucedropemoff",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$ktitle brucedropemoff"},
			APIResp:    kicktest.LargeLiveGetChannelResp,
			Want: []*base.Message{
				{
					Text:    "brucedropemoff's title on Kick: DEO COOKOFF MAY THE BEST DISH WIN! 🗣️ #DEO4L",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kicktitle xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$ktitle xqc"},
			APIResp:    kicktest.LargeOfflineGetChannelResp,
			Want: []*base.Message{
				{
					Text:    "Currently Kick only returns the title for live channels, and xqc is not currently live.",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}
