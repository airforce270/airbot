// Package admin handles bot administration commands.
package admin

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/utils"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	botSlowmodeCommand,
	echoCommand,
	joinCommand,
	joinOtherCommand,
	joinedCommand,
	leaveCommand,
	leaveOtherCommand,
	setPrefixCommand,
}

var (
	botSlowmodeCommand = basecommand.Command{
		Name:       "botslowmode",
		Help:       "Sets the bot to follow a global (per-platform) 1 second slowmode. If no argument is provided, checks if slowmode is enabled.",
		Args:       []basecommand.Argument{{Name: "enable", Required: false, Usage: "on|off"}},
		Permission: permission.Owner,
		Handler:    botSlowmode,
	}

	echoCommand = basecommand.Command{
		Name:       "echo",
		Help:       "Echoes back whatever is sent.",
		Args:       generateEchoArgs(),
		Permission: permission.Owner,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			if len(args) == 0 {
				return nil, basecommand.ErrBadUsage
			}
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    strings.Join(args, " "),
				},
			}, nil
		},
	}

	joinCommand = basecommand.Command{
		Name:       "join",
		Help:       "Tells the bot to join your chat.",
		Args:       []basecommand.Argument{{Name: "prefix", Required: false}},
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			prefix := defaultPrefix
			if len(args) >= 1 {
				prefix = args[0]
			}
			return joinChannel(msg, msg.Message.User, prefix)
		},
	}

	joinedCommand = basecommand.Command{
		Name:       "joined",
		Help:       "Lists the channels the bot is currently in.",
		Permission: permission.Owner,
		Handler:    joined,
	}

	joinOtherCommand = basecommand.Command{
		Name: "joinother",
		Help: "Tells the bot to join a chat.",
		Args: []basecommand.Argument{
			{Name: "channel", Required: true},
			{Name: "prefix", Required: false},
		},
		Permission: permission.Owner,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			if len(args) == 0 {
				return nil, basecommand.ErrBadUsage
			}
			channel := args[0]
			prefix := defaultPrefix
			if len(args) >= 2 {
				prefix = args[1]
			}
			return joinChannel(msg, channel, prefix)
		},
	}

	leaveCommand = basecommand.Command{
		Name:       "leave",
		Help:       "Tells the bot to leave your chat.",
		Permission: permission.Admin,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			return leaveChannel(msg, msg.Message.Channel)
		},
	}

	leaveOtherCommand = basecommand.Command{
		Name:       "leaveother",
		Help:       "Tells the bot to leave a chat.",
		Args:       []basecommand.Argument{{Name: "channel", Required: true}},
		Permission: permission.Owner,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			if len(args) == 0 {
				return nil, basecommand.ErrBadUsage
			}
			return leaveChannel(msg, args[0])
		},
	}

	setPrefixCommand = basecommand.Command{
		Name:       "setprefix",
		Help:       "Sets the bot's prefix in the channel.",
		Args:       []basecommand.Argument{{Name: "prefix", Required: true}},
		Permission: permission.Admin,
		Handler:    setPrefix,
	}
)

func botSlowmode(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	cdb := cache.Instance
	if cdb == nil {
		return nil, fmt.Errorf("cache instance not initialized")
	}

	if len(args) == 0 {
		enabled, err := cdb.FetchBool(cache.KeyGlobalSlowmode(msg.Platform))
		if err != nil {
			return nil, err
		}
		enabledMsg := "enabled"
		if !enabled {
			enabledMsg = "disabled"
		}
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Bot slowmode is currently %s on %s", enabledMsg, msg.Platform.Name()),
			},
		}, nil
	}
	if args[0] != "on" && args[0] != "off" {
		return nil, nil
	}
	enable := args[0] == "on"

	err := cdb.StoreBool(cache.KeyGlobalSlowmode(msg.Platform), enable)
	if err != nil {
		failureMsgStart := "Failed to enable"
		if !enable {
			failureMsgStart = "Failed to disable"
		}
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s bot slowmode on %s", failureMsgStart, msg.Platform.Name()),
			},
		}, nil
	}

	outMsgStart := "Enabled"
	if !enable {
		outMsgStart = "Disabled"
	}
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s bot slowmode on %s", outMsgStart, msg.Platform.Name()),
		},
	}, nil
}

const defaultPrefix = "$"

func joinChannel(msg *base.IncomingMessage, targetChannel, prefix string) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database instance not initialized")
	}

	var channels []models.JoinedChannel
	db.Where(models.JoinedChannel{
		Platform: msg.Platform.Name(),
		Channel:  strings.ToLower(targetChannel),
	}).Find(&channels)

	if len(channels) > 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Channel %s is already joined", targetChannel),
			},
		}, nil
	}

	channelRecord := models.JoinedChannel{
		Platform: msg.Platform.Name(),
		Channel:  targetChannel,
		Prefix:   prefix,
		JoinedAt: time.Now(),
	}
	result := db.Create(&channelRecord)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to join channel %s: %w", targetChannel, result.Error)
	}

	err := msg.Platform.Join(targetChannel, prefix)

	if errors.Is(err, twitchplatform.ErrChannelNotFound) {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Channel %s not found", targetChannel),
			},
		}, nil
	}
	if errors.Is(err, twitchplatform.ErrBotIsBanned) {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Bot is banned from %s", targetChannel),
			},
		}, nil
	}

	msgs := []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Successfully joined channel %s with prefix %s", targetChannel, prefix),
		},
	}
	if !strings.EqualFold(msg.Message.Channel, targetChannel) {
		msgs = append(msgs, &base.Message{
			Channel: targetChannel,
			Text:    fmt.Sprintf("Successfully joined channel! (prefix: %s ) For all commands, type %scommands.", prefix, prefix),
		})
	}
	return msgs, nil
}

const maxUsersPerMessage = 15

func joined(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database instance not initialized")
	}

	var joinedChannels []*models.JoinedChannel
	result := db.Find(&joinedChannels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find channels: %w", result.Error)
	}
	var channels []string
	for _, c := range joinedChannels {
		channels = append(channels, c.Channel)
	}

	channelsGroups := utils.Chunk(channels, maxUsersPerMessage)

	var messages []*base.Message

	for i, channelsGroup := range channelsGroups {
		var text string
		if i == 0 {
			text = fmt.Sprintf("Bot is currently in %s", strings.Join(channelsGroup, ", "))
		} else {
			text = strings.Join(channels, ", ")
		}
		if len(channelsGroups) > 1 && len(channelsGroups)-1 != i {
			text += ","
		}
		messages = append(messages, &base.Message{Channel: msg.Message.Channel, Text: text})
	}

	return messages, nil
}

func leaveChannel(msg *base.IncomingMessage, targetChannel string) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database instance not initialized")
	}

	err := database.LeaveChannel(db, msg.Platform.Name(), targetChannel)

	if err != nil {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Bot is not in channel %s", targetChannel),
			},
		}, nil
	}

	go func() {
		time.Sleep(time.Millisecond * 500)
		if err := msg.Platform.Leave(targetChannel); err != nil {
			log.Printf("failed to leave channel %s: %v", targetChannel, err)
		}
	}()

	var msgs []*base.Message
	if strings.EqualFold(msg.Message.Channel, targetChannel) {
		msgs = append(msgs, &base.Message{
			Channel: msg.Message.Channel,
			Text:    "Successfully left channel.",
		})
	} else {
		msgs = append(msgs, &base.Message{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Successfully left channel %s", targetChannel),
		})
	}
	return msgs, nil
}

func setPrefix(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) == 0 {
		return nil, basecommand.ErrBadUsage
	}
	newPrefix := args[0]

	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database instance not initialized")
	}

	var channels []models.JoinedChannel
	db.Where("platform = ? AND LOWER(channel) = ?", msg.Platform.Name(), strings.ToLower(msg.Message.Channel)).Find(&channels)

	for _, channel := range channels {
		channel.Prefix = newPrefix

		result := db.Save(&channel)

		if result.RowsAffected == 0 {
			log.Printf("Failed to update prefix: %v", result.Error)
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Failed to update prefix",
				},
			}, nil
		}
	}

	if err := msg.Platform.SetPrefix(msg.Message.Channel, newPrefix); err != nil {
		log.Printf("Failed to update prefix: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Failed to update prefix",
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Prefix set to %s", newPrefix),
		},
	}, nil
}

// Generate args for echo command, which is special as the number of args isn't limited.
func generateEchoArgs() []basecommand.Argument {
	var args [1000]basecommand.Argument
	args[0] = basecommand.Argument{
		Name:     "arg-0",
		Required: true,
		Usage:    "anything",
	}
	for i := 1; i < 1000; i++ {
		args[i] = basecommand.Argument{
			Name:             fmt.Sprintf("arg-%d", i),
			ExcludeFromUsage: true,
		}
	}
	return args[:]
}
