// Package commands contains all commands that can be run and a command handler.
package commands

import (
	"strings"

	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/commands/echo"
	"github.com/airforce270/airbot/commands/twitch"
	"github.com/airforce270/airbot/message"
)

// allCommands contains all allCommands that can be run.
var allCommands []basecommand.Command

// Handler handles messages.
type Handler struct{}

// Handle handles incoming messages, possibly returning messages to be sent in response.
func (h *Handler) Handle(msg *message.IncomingMessage) ([]*message.Message, error) {
	var outMsgs []*message.Message
	for _, command := range allCommands {
		if command.PrefixOnly && !strings.HasPrefix(msg.Message.Text, msg.Prefix) {
			continue
		}
		if !command.Pattern.MatchString(msg.MessageTextWithoutPrefix()) {
			continue
		}

		respMsgs, err := command.Handle(msg)
		if err != nil {
			return nil, err
		}

		outMsgs = append(outMsgs, respMsgs...)
	}
	return outMsgs, nil
}

func init() {
	allCommands = append(allCommands, echo.Commands[:]...)
	allCommands = append(allCommands, twitch.Commands[:]...)
}
