// Package commands contains all commands that can be run and a command handler.
package commands

import (
	"strings"

	"github.com/airforce270/airbot/commands/admin"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/commands/botinfo"
	"github.com/airforce270/airbot/commands/echo"
	"github.com/airforce270/airbot/commands/twitch"
	"github.com/airforce270/airbot/message"

	"golang.org/x/exp/slices"
)

// allCommands contains all allCommands that can be run.
var allCommands []basecommand.Command

// NewHandler creates a new Handler.
func NewHandler(enableNonPrefixCommands bool, admins []string) Handler {
	return Handler{
		nonPrefixCommandsEnabled: enableNonPrefixCommands,
		admins:                   admins,
	}
}

// Handler handles messages.
type Handler struct {
	// nonPrefixCommandsEnabled is whether non-prefix commands should be enabled.
	nonPrefixCommandsEnabled bool
	// admins is the list of admins.
	admins []string
}

// Handle handles incoming messages, possibly returning messages to be sent in response.
func (h *Handler) Handle(msg *message.IncomingMessage) ([]*message.Message, error) {
	var outMsgs []*message.Message
	for _, command := range allCommands {
		messageHasPrefix := strings.HasPrefix(msg.Message.Text, msg.Prefix)
		if !messageHasPrefix && (command.PrefixOnly || !h.nonPrefixCommandsEnabled) {
			continue
		}
		if command.AdminOnly && !slices.Contains(h.admins, msg.Message.User) {
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
	allCommands = append(allCommands, admin.Commands[:]...)
	allCommands = append(allCommands, botinfo.Commands[:]...)
	allCommands = append(allCommands, echo.Commands[:]...)
	allCommands = append(allCommands, twitch.Commands[:]...)
}
