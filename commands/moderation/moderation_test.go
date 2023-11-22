package moderation_test

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/permission"
)

func TestModerationCommands(t *testing.T) {
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vanish",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunAfter: []commandtest.TeardownFunc{
				waitForMessagesToSend,
			},
			Want: nil,
		},
	}

	commandtest.Run(t, tests)
}

func waitForMessagesToSend(t testing.TB) {
	time.Sleep(20 * time.Millisecond)
}
