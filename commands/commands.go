// Package commands contains all commands that can be run and a command handler.
package commands

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/ivr"
	kickapi "github.com/airforce270/airbot/apiclients/kick"
	seventvapi "github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/commands/admin"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/commands/botinfo"
	"github.com/airforce270/airbot/commands/bulk"
	"github.com/airforce270/airbot/commands/echo"
	"github.com/airforce270/airbot/commands/fun"
	"github.com/airforce270/airbot/commands/gamba"
	"github.com/airforce270/airbot/commands/kick"
	"github.com/airforce270/airbot/commands/moderation"
	"github.com/airforce270/airbot/commands/seventv"
	"github.com/airforce270/airbot/commands/twitch"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"

	"gorm.io/gorm"
)

// CommandGroups contains all groups of commands.
var CommandGroups = map[string][]basecommand.Command{
	"7TV":        seventv.Commands[:],
	"Admin":      admin.Commands[:],
	"Bot info":   append([]basecommand.Command{helpCommand}, botinfo.Commands[:]...),
	"Bulk":       bulk.Commands[:],
	"Fun":        fun.Commands[:],
	"Gamba":      gamba.Commands[:],
	"Kick":       kick.Commands[:],
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
func NewHandler(ctx context.Context, db *gorm.DB, cdb cache.Cache, cfg *config.Config, allPlatforms map[string]base.Platform) Handler {
	return Handler{
		db:              db,
		cache:           cdb,
		allPlatforms:    allPlatforms,
		newConfigSource: config.DefaultNewConfigSource,
		rand: base.RandResources{
			Reader: rand.Reader,
		},
		Clients: base.APIClients{
			Bible:   bible.NewDefaultClient(),
			IVR:     ivr.NewDefaultClient(),
			Kick:    kickapi.NewClient(kickapi.DefaultBaseURL, cfg.Platforms.Kick.JA3, cfg.Platforms.Kick.UserAgent),
			SevenTV: seventvapi.NewClient(ctx, seventvapi.DefaultBaseURL, cfg.SevenTV.AccessToken),
		},
	}
}

// NewHandlerForTest creates a new Handler for use in testing.
func NewHandlerForTest(db *gorm.DB, cdb cache.Cache, allPlatforms map[string]base.Platform, newConfigSource func() (io.ReadCloser, error), randOpts base.RandResources, clients base.APIClients) Handler {
	return Handler{
		db:              db,
		cache:           cdb,
		allPlatforms:    allPlatforms,
		newConfigSource: newConfigSource,
		rand:            randOpts,
		Clients:         clients,
	}
}

// Handler handles messages.
type Handler struct {
	// db is a connection to the database.
	db *gorm.DB
	// cache is a reference to the cache.
	cache cache.Cache
	// allPlatforms contains all registered and configured platforms.
	allPlatforms map[string]base.Platform
	// newConfigSource returns a source for the latest config data.
	newConfigSource func() (io.ReadCloser, error)
	// rand are references to sources of randomness.
	rand base.RandResources
	// Clients are API Clients to use.
	// It is only exported for testing and should not be otherwise used.
	Clients base.APIClients
}

// Handle handles incoming messages, possibly returning messages to be sent in response.
func (h *Handler) Handle(msg *base.IncomingMessage) ([]*base.OutgoingMessage, error) {
	h.setResources(msg)

	var outMsgs []*base.OutgoingMessage
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
		shouldSetChannelCooldown := true
		err := h.db.FirstOrCreate(&channelCooldown, models.ChannelCommandCooldown{
			Channel: msg.Message.Channel,
			Command: command.Name,
		}).Error
		if err != nil {
			return nil, fmt.Errorf("[%s] failed to get/create channel cooldown for channel %q, command %q: %w", msg.Resources.Platform.Name(), msg.Message.Channel, command.Name, err)
		}
		if command.ChannelCooldown > time.Since(channelCooldown.LastRun) {
			log.Printf("Skipping %s%s: channel cooldown is %s but it has only been %s", msg.Prefix, command.Name, command.ChannelCooldown, time.Since(channelCooldown.LastRun))
			continue
		}

		user, err := msg.Resources.Platform.User(msg.Message.User)
		if err != nil && !errors.Is(err, base.ErrUserUnknown) {
			return nil, fmt.Errorf("failed to fetch user %q: %w", msg.Message.User, err)
		}
		userCooldown := models.UserCommandCooldown{}
		shouldSetUserCooldown := false
		if err == nil || errors.Is(err, base.ErrUserUnknown) {
			shouldSetUserCooldown = true
			err := h.db.FirstOrCreate(&userCooldown, models.UserCommandCooldown{
				UserID:  user.ID,
				User:    user,
				Command: command.Name,
			}).Error
			if err != nil {
				return nil, fmt.Errorf("[%s] failed to get/create user cooldown for user %q, command %q, %w", msg.Resources.Platform.Name(), msg.Message.User, command.Name, err)
			}
			if command.UserCooldown > time.Since(userCooldown.LastRun) {
				log.Printf("Skipping %s%s: user cooldown is %s but it has only been %s", msg.Prefix, command.Name, command.UserCooldown, time.Since(userCooldown.LastRun))
				continue
			}
		}

		args := parseArgs(msg, command, pattern)

		respMsgs, err := command.Handler(msg, args)
		if err != nil {
			if !errors.Is(err, basecommand.ErrBadUsage) {
				return nil, fmt.Errorf("failed to handle message: %w", err)
			}
			shouldSetChannelCooldown = false
			outMsg := &base.OutgoingMessage{
				Message: base.Message{
					Channel: msg.Message.Channel,
					Text:    "Usage: " + command.Usage(msg.Prefix),
				},
			}
			if !command.DisableReplies {
				outMsg.ReplyToID = msg.Message.ID
			}
			outMsgs = append(outMsgs, outMsg)
		} else {
			for _, respMsg := range respMsgs {
				outMsg := &base.OutgoingMessage{Message: *respMsg}
				if !command.DisableReplies && respMsg.Channel == msg.Message.Channel {
					outMsg.ReplyToID = msg.Message.ID
				}
				outMsgs = append(outMsgs, outMsg)
			}
		}

		if shouldSetChannelCooldown {
			channelCooldown.LastRun = time.Now()
			if err := h.db.Save(&channelCooldown).Error; err != nil {
				log.Printf("failed to save new channel cooldown: %v", err)
			}
		}
		if shouldSetUserCooldown {
			userCooldown.LastRun = time.Now()
			if err := h.db.Save(&userCooldown).Error; err != nil {
				log.Printf("failed to save new user cooldown: %v", err)
			}
		}
	}
	return outMsgs, nil
}

func (h *Handler) setResources(msg *base.IncomingMessage) {
	msg.Resources.DB = h.db
	msg.Resources.Cache = h.cache
	// msg.Resources.Platform is set by the platform,
	// before the message reaches here
	msg.Resources.AllPlatforms = h.allPlatforms
	msg.Resources.NewConfigSource = h.newConfigSource
	msg.Resources.Rand = h.rand
	msg.Resources.Clients = h.Clients
}

// parseArgs parses all args from the given message.
func parseArgs(msg *base.IncomingMessage, cmd basecommand.Command, pattern *regexp.Regexp) []arg.Arg {
	var parsed []arg.Arg
	rest := strings.TrimSpace(strings.Join(pattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())[1:], " "))

	for _, a := range cmd.Params {
		var value arg.Arg
		value, rest = a.Parse(rest)
		parsed = append(parsed, value)
	}

	return parsed
}

var (
	helpCommand = basecommand.Command{
		Name:    "help",
		Desc:    "Displays help for a command.",
		Params:  []arg.Param{{Name: "command", Type: arg.String, Required: false}},
		Handler: help,
	}
)

func help(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetCommandArg := args[0]
	if !targetCommandArg.Present {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("For help with a command, use %shelp <command>. To see available commands, use %scommands", msg.Prefix, msg.Prefix),
			},
		}, nil
	}
	targetCommand := targetCommandArg.StringValue

	for _, cmd := range allCommands {
		if !strings.EqualFold(cmd.Name, targetCommand) {
			continue
		}
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("[ %s%s ] %s", msg.Prefix, cmd.Name, cmd.Help()),
			},
		}, nil
	}

	return nil, nil
}
