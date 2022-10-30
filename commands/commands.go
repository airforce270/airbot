// Package commands contains all commands that can be run and a command handler.
package commands

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/admin"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/commands/botinfo"
	"github.com/airforce270/airbot/commands/bulk"
	"github.com/airforce270/airbot/commands/echo"
	"github.com/airforce270/airbot/commands/twitch"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"

	"gorm.io/gorm"
)

// CommandGroups contains all groups of commands.
var CommandGroups = map[string][]basecommand.Command{
	"Admin":    admin.Commands[:],
	"Bot info": append([]basecommand.Command{helpCommand}, botinfo.Commands[:]...),
	"Bulk":     bulk.Commands[:],
	"Echo":     echo.Commands[:],
	"Twitch":   twitch.Commands[:],
}

// allCommands contains all all commands that can be run.
var allCommands []basecommand.Command

func init() {
	for _, group := range CommandGroups {
		allCommands = append(allCommands, group...)
	}
}

// NewHandler creates a new Handler.
func NewHandler(db *gorm.DB, enableNonPrefixCommands bool) Handler {
	return Handler{
		db:                       db,
		nonPrefixCommandsEnabled: enableNonPrefixCommands,
	}
}

// Handler handles messages.
type Handler struct {
	// db is a connection to the database.
	db *gorm.DB
	// nonPrefixCommandsEnabled is whether non-prefix commands should be enabled.
	nonPrefixCommandsEnabled bool
}

// Handle handles incoming messages, possibly returning messages to be sent in response.
func (h *Handler) Handle(msg *base.IncomingMessage) ([]*base.Message, error) {
	var outMsgs []*base.Message
	for _, command := range allCommands {
		messageHasPrefix := strings.HasPrefix(msg.Message.Text, msg.Prefix)
		if !messageHasPrefix && (command.PrefixOnly || !h.nonPrefixCommandsEnabled) {
			continue
		}
		if !command.Pattern.MatchString(msg.MessageTextWithoutPrefix()) {
			continue
		}
		if !permission.Authorized(msg.PermissionLevel, command.Permission) {
			log.Printf("Permission denied: command %s, user %s, channel %s; has permission %s, required: %s", command.Name, msg.Message.User, msg.Message.Channel, msg.PermissionLevel.Name(), command.Permission.Name())
			continue
		}
		channelCooldown := models.ChannelCommandCooldown{}
		h.db.FirstOrCreate(&channelCooldown, models.ChannelCommandCooldown{
			Channel: msg.Message.Channel,
			Command: command.Name,
		})
		if command.ChannelCooldown > time.Since(channelCooldown.LastRun) {
			log.Printf("Skipping %s%s: cooldown is %s but it has only been %s", msg.Prefix, command.Name, command.ChannelCooldown, time.Since(channelCooldown.LastRun))
			continue
		}

		respMsgs, err := command.Handler(msg)
		if err != nil {
			return nil, err
		}

		outMsgs = append(outMsgs, respMsgs...)

		channelCooldown.LastRun = time.Now()
		h.db.Save(&channelCooldown)
	}
	return outMsgs, nil
}

var (
	helpCommandPattern = basecommand.PrefixPattern("help")
	helpCommand        = basecommand.Command{
		Name:       "help",
		Help:       "Displays help for a command.",
		Usage:      "$help <command>",
		Pattern:    helpCommandPattern,
		Handler:    help,
		PrefixOnly: true,
	}
	helpPattern = regexp.MustCompile(helpCommandPattern.String() + `(\w+).*`)
)

func help(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetCommand := basecommand.ParseTarget(msg, helpPattern)

	// No command provided
	if targetCommand == msg.Message.User {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("For help with a command, use %shelp <command>. To see available commands, use %scommands", msg.Prefix, msg.Prefix),
			},
		}, nil
	}

	for _, cmd := range allCommands {
		if !strings.EqualFold(cmd.Name, targetCommand) {
			continue
		}
		help := cmd.Help
		if help == "" {
			help = "<no help information found>"
		}
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("[ %s%s ] %s", msg.Prefix, cmd.Name, help),
			},
		}, nil
	}

	return nil, nil
}
