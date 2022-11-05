// Package moderation implements moderation commands.
package moderation

import (
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	vanishCommand,
}

var (
	vanishCommandPattern = basecommand.PrefixPattern(`vanish`)
	vanishCommand        = basecommand.Command{
		Name:       "vanish",
		Help:       "Times you out for 1 second.",
		Usage:      "$vanish",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    vanishCommandPattern,
		Handler:    vanish,
	}
)

func vanish(msg *base.IncomingMessage) ([]*base.Message, error) {
	if err := msg.Platform.Timeout(msg.Message.Channel, msg.Message.User, time.Duration(1)*time.Second); err != nil {
		return nil, err
	}
	return nil, nil
}
