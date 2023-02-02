// Package moderation implements moderation commands.
package moderation

import (
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
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
		Desc:       "Times you out for 1 second.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			err := msg.Platform.Timeout(msg.Message.User, msg.Message.Channel, time.Duration(1)*time.Second)
			return nil, err
		},
	}
)
