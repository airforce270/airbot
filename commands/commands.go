// Package commands contains all commands that can be run and a command handler.
package commands

import (
	"errors"
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
	"github.com/airforce270/airbot/commands/fun"
	"github.com/airforce270/airbot/commands/gamba"
	"github.com/airforce270/airbot/commands/moderation"
	"github.com/airforce270/airbot/commands/twitch"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"

	"gorm.io/gorm"
)

// CommandGroups contains all groups of commands.
var CommandGroups = map[string][]basecommand.Command{
	"Admin":      admin.Commands[:],
	"Bot info":   append([]basecommand.Command{helpCommand}, botinfo.Commands[:]...),
	"Bulk":       bulk.Commands[:],
	"Fun":        fun.Commands[:],
	"Gamba":      gamba.Commands[:],
	"Moderation": moderation.Commands[:],
	"Echo":       echo.Commands[:],
	"Twitch":     twitch.Commands[:],
}

var (
	// allCommands contains all commands that can be run.
	allCommands []basecommand.Command
	// commandPatterns contains a map of patterns to trigger a command to that command.
	commandPatterns = map[*regexp.Regexp]basecommand.Command{}
)

func init() {
	for _, group := range CommandGroups {
		allCommands = append(allCommands, group...)
	}
	for _, command := range allCommands {
		pattern, err := command.Compile()
		if err != nil {
			panic(fmt.Sprintf("failed to compile pattern for %s: %v", command.Name, err))
		}
		commandPatterns[pattern] = command
	}
}

// NewHandler creates a new Handler.
func NewHandler(db *gorm.DB) Handler {
	return Handler{db: db}
}

// Handler handles messages.
type Handler struct {
	// db is a connection to the database.
	db *gorm.DB
}

// Handle handles incoming messages, possibly returning messages to be sent in response.
func (h *Handler) Handle(msg *base.IncomingMessage) ([]*base.Message, error) {
	var outMsgs []*base.Message
	for pattern, command := range commandPatterns {
		if !strings.HasPrefix(strings.TrimSpace(msg.Message.Text), msg.Prefix) {
			continue
		}
		if !pattern.MatchString(msg.MessageTextWithoutPrefix()) {
			continue
		}
		if !permission.Authorized(msg.PermissionLevel, command.Permission) {
			log.Printf("Permission denied: command %s, user %s, channel %s; has permission %s, required: %s", command.Name, msg.Message.User, msg.Message.Channel, msg.PermissionLevel.Name(), command.Permission.Name())
			continue
		}

		channelCooldown := models.ChannelCommandCooldown{}
		dbResult := h.db.FirstOrCreate(&channelCooldown, models.ChannelCommandCooldown{
			Channel: msg.Message.Channel,
			Command: command.Name,
		})
		if dbResult.Error != nil {
			return nil, fmt.Errorf("[%s] failed to get/create channel cooldown for channel %q, command %q", msg.Platform.Name(), msg.Message.Channel, command.Name)
		}
		if command.ChannelCooldown > time.Since(channelCooldown.LastRun) {
			log.Printf("Skipping %s%s: channel cooldown is %s but it has only been %s", msg.Prefix, command.Name, command.ChannelCooldown, time.Since(channelCooldown.LastRun))
			continue
		}

		user, err := msg.Platform.User(msg.Message.User)
		if err != nil && !errors.Is(err, base.ErrUserUnknown) {
			return nil, fmt.Errorf("failed to fetch user %q: %v", msg.Message.User, err)
		}
		userCooldown := models.UserCommandCooldown{}
		shouldSetUserCooldown := false
		if err == nil || errors.Is(err, base.ErrUserUnknown) {
			shouldSetUserCooldown = true
			dbResult = h.db.FirstOrCreate(&userCooldown, models.UserCommandCooldown{
				UserID:  user.ID,
				User:    user,
				Command: command.Name,
			})
			if dbResult.Error != nil {
				return nil, fmt.Errorf("[%s] failed to get/create user cooldown for user %q, command %q", msg.Platform.Name(), msg.Message.User, command.Name)
			}
			if command.UserCooldown > time.Since(userCooldown.LastRun) {
				log.Printf("Skipping %s%s: user cooldown is %s but it has only been %s", msg.Prefix, command.Name, command.UserCooldown, time.Since(userCooldown.LastRun))
				continue
			}
		}

		args := parseArgs(pattern, msg)

		respMsgs, err := command.Handler(msg, args)
		if err != nil {
			if !errors.Is(err, basecommand.ErrReturnUsage) {
				return nil, err
			}
			outMsgs = append(outMsgs, &base.Message{
				Channel: msg.Message.Channel,
				Text:    "Usage: " + command.Usage(msg.Prefix),
			})
		} else {
			outMsgs = append(outMsgs, respMsgs...)
		}

		channelCooldown.LastRun = time.Now()
		h.db.Save(&channelCooldown)
		if shouldSetUserCooldown {
			userCooldown.LastRun = time.Now()
			h.db.Save(&userCooldown)
		}
	}
	return outMsgs, nil
}

func parseArgs(pattern *regexp.Regexp, msg *base.IncomingMessage) []string {
	var args []string
	for _, arg := range pattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())[1:] {
		if arg == "" {
			continue
		}
		args = append(args, arg)
	}
	return args
}

var (
	helpCommand = basecommand.Command{
		Name:    "help",
		Help:    "Displays help for a command.",
		Args:    []basecommand.Argument{{Name: "command", Required: false}},
		Handler: help,
	}
)

func help(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("For help with a command, use %shelp <command>. To see available commands, use %scommands", msg.Prefix, msg.Prefix),
			},
		}, nil
	}
	targetCommand := args[0]

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
