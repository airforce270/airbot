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
	tests := []commandtest.Case{
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
			ApiResps: []string{
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
			ApiResps: []string{
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
	}

	commandtest.Run(t, tests)
}
