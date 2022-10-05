// Package commands contains all commands that can be run and a command handler.
package commands

import (
	"strings"

	"airbot/commands/basecommand"
	"airbot/commands/echo"
	"airbot/message"
)

// allCommands contains all allCommands that can be run.
var allCommands []basecommand.Command

// Handler handles messages.
type Handler struct {
	// Username is the bot's username.
	Username string
}

// Handle handles incoming messages, possibly returning a message to be sent in response.
func (h *Handler) Handle(msg *message.Message) (*message.Message, error) {
	for _, command := range allCommands {
		if !strings.HasPrefix(msg.Text, command.Prefix) {
			continue
		}
		return command.F(msg)
	}

	return nil, nil
}

func init() {
	allCommands = append(allCommands, echo.Commands...)
}
