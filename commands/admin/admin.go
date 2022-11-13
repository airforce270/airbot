// Package admin handles bot administration commands.
package admin

import (
	"errors"
	"fmt"
	"log"
	"regexp"
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
	botSlowmodeCommandPattern = basecommand.PrefixPattern("botslowmode")
	botSlowmodeCommand        = basecommand.Command{
		Name:       "botslowmode",
		Help:       "Sets the bot to follow a global (per-platform) 1 second slowmode.",
		Usage:      "$botslowmode <on|off>",
		Permission: permission.Owner,
		PrefixOnly: true,
		Pattern:    botSlowmodeCommandPattern,
		Handler:    botSlowmode,
	}
	botSlowmodePattern = regexp.MustCompile(botSlowmodeCommandPattern.String() + `(on|off).*`)

	echoCommandPattern = basecommand.PrefixPattern("echo")
	echoCommand        = basecommand.Command{
		Name:       "echo",
		Help:       "Echoes back whatever is sent.",
		Usage:      "$echo",
		Permission: permission.Owner,
		PrefixOnly: true,
		Pattern:    echoCommandPattern,
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
			matches := echoPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
			if len(matches) < 2 {
				return nil, nil
			}
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    matches[1],
				},
			}, nil
		},
	}
	echoPattern = regexp.MustCompile(echoCommandPattern.String() + `(.+)`)

	joinCommandPattern = basecommand.PrefixPattern("join$")
	joinCommand        = basecommand.Command{
		Name:       "join",
		Help:       "Tells the bot to join your chat.",
		Usage:      "$join",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    joinCommandPattern,
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
			return joinChannel(msg, msg.Message.User)
		},
	}

	joinedCommandPattern = basecommand.PrefixPattern("joined$")
	joinedCommand        = basecommand.Command{
		Name:       "joined",
		Help:       "Lists the channels the bot is currently in.",
		Usage:      "$joined",
		Permission: permission.Owner,
		PrefixOnly: true,
		Pattern:    joinedCommandPattern,
		Handler:    joined,
	}

	joinOtherCommandPattern = basecommand.PrefixPattern("joinother")
	joinOtherCommand        = basecommand.Command{
		Name:       "joinother",
		Help:       "Tells the bot to join a chat.",
		Usage:      "$joinother <channel>",
		Permission: permission.Owner,
		PrefixOnly: true,
		Pattern:    joinOtherCommandPattern,
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
			return joinChannel(msg, basecommand.ParseTarget(msg, joinOtherPattern))
		},
	}
	joinOtherPattern = regexp.MustCompile(joinOtherCommandPattern.String() + `(\w+).*`)

	leaveCommandPattern = basecommand.PrefixPattern("leave$")
	leaveCommand        = basecommand.Command{
		Name:       "leave",
		Help:       "Tells the bot to leave your chat.",
		Usage:      "$leave",
		Permission: permission.Admin,
		PrefixOnly: true,
		Pattern:    leaveCommandPattern,
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
			return leaveChannel(msg, msg.Message.Channel)
		},
	}

	leaveOtherCommandPattern = basecommand.PrefixPattern("leaveother")
	leaveOtherCommand        = basecommand.Command{
		Name:       "leaveother",
		Help:       "Tells the bot to leave a chat.",
		Usage:      "$leaveother <channel>",
		Permission: permission.Owner,
		PrefixOnly: true,
		Pattern:    leaveOtherCommandPattern,
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
			return leaveChannel(msg, basecommand.ParseTarget(msg, leaveOtherPattern))
		},
	}
	leaveOtherPattern = regexp.MustCompile(leaveOtherCommandPattern.String() + `(\w+).*`)

	setPrefixCommandPattern = basecommand.PrefixPattern("setprefix")
	setPrefixCommand        = basecommand.Command{
		Name:       "setprefix",
		Help:       "Sets the bot's prefix in the channel.",
		Usage:      "$setprefix",
		Permission: permission.Admin,
		PrefixOnly: true,
		Pattern:    setPrefixCommandPattern,
		Handler:    setPrefix,
	}
	setPrefixPattern = regexp.MustCompile(setPrefixCommandPattern.String() + `(\S+).*`)
)

func botSlowmode(msg *base.IncomingMessage) ([]*base.Message, error) {
	cdb := cache.Instance
	if cdb == nil {
		return nil, fmt.Errorf("cache instance not initialized")
	}

	matches := botSlowmodePattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) < 2 {
		enabled, err := cache.FetchSlowmode(msg.Platform, cdb)
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
	if matches[1] != "on" && matches[1] != "off" {
		return nil, nil
	}
	enable := matches[1] == "on"

	err := cache.SetSlowmode(msg.Platform, cdb, enable)
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

func joinChannel(msg *base.IncomingMessage, targetChannel string) ([]*base.Message, error) {
	prefix := defaultPrefix

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

func joined(msg *base.IncomingMessage) ([]*base.Message, error) {
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

func setPrefix(msg *base.IncomingMessage) ([]*base.Message, error) {
	matches := setPrefixPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) < 2 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "No new prefix provided",
			},
		}, nil
	}
	newPrefix := matches[1]

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

	msg.Platform.SetPrefix(msg.Message.Channel, newPrefix)

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Prefix set to %s", newPrefix),
		},
	}, nil
}
