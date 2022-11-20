// Package moderation implements moderation commands.
package moderation

import (
	"fmt"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	vanishCommand,
}

var (
	vanishCommand = basecommand.Command{
		Name:       "vanish",
		Help:       "Times you out for 1 second.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("/timeout %s 1", msg.Message.User),
				},
			}, nil
		},
	}
)
