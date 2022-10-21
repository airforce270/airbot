// Package admin handles bot administration commands.
package admin

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/model"
	"github.com/airforce270/airbot/message"
	"github.com/airforce270/airbot/permission"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	joinCommand,
	joinOtherCommand,
	leaveCommand,
	leaveOtherCommand,
	setPrefixCommand,
}

var (
	joinCommandPattern = basecommand.PrefixPattern("join$")
	joinCommand        = basecommand.Command{
		Name:       "join",
		Help:       "Tells the bot to join your chat.",
		Usage:      "$join",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    joinCommandPattern,
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
			return joinChannel(msg, msg.Message.User)
		},
	}

	joinOtherCommandPattern = basecommand.PrefixPattern("joinother")
	joinOtherCommand        = basecommand.Command{
		Name:       "joinother",
		Help:       "Tells the bot to join a chat.",
		Usage:      "$joinother <channel>",
		Permission: permission.Owner,
		PrefixOnly: true,
		Pattern:    joinOtherCommandPattern,
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
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
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
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
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
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

const defaultPrefix = "$"

func joinChannel(msg *message.IncomingMessage, targetChannel string) ([]*message.Message, error) {
	prefix := defaultPrefix

	db := database.Instance

	var channels []model.JoinedChannel
	db.Where(model.JoinedChannel{Platform: "Twitch", Channel: strings.ToLower(targetChannel)}).Find(&channels)

	if len(channels) > 0 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Channel %s is already joined", targetChannel),
			},
		}, nil
	}

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}
	channel, err := tw.Channel(targetChannel)
	if err != nil {
		if errors.Is(err, twitchplatform.ErrChannelNotFound) {
			return []*message.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("Channel %s not found", targetChannel),
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to look up channel: %w", err)
	}

	channelRecord := model.JoinedChannel{
		Platform: "Twitch",
		Channel:  channel.BroadcasterName,
		Prefix:   prefix,
		JoinedAt: time.Now(),
	}
	result := db.Create(&channelRecord)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to join channel %s: %w", channel.BroadcasterName, err)
	}

	if err := tw.Join(channel, prefix); err != nil {
		return nil, fmt.Errorf("failed to join channel %s: %w", channel.BroadcasterName, err)
	}

	msgs := []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Successfully joined channel %s with prefix %s", channel.BroadcasterName, prefix),
		},
	}
	if !strings.EqualFold(msg.Message.Channel, channel.BroadcasterName) {
		msgs = append(msgs, &message.Message{
			Channel: channel.BroadcasterName,
			Text:    fmt.Sprintf("Successfully joined channel! (prefix: %s ) For all commands, type %scommands.", prefix, prefix),
		})
	}
	return msgs, nil
}

func leaveChannel(msg *message.IncomingMessage, targetChannel string) ([]*message.Message, error) {
	db := database.Instance

	var channels []model.JoinedChannel
	db.Where(model.JoinedChannel{Platform: "Twitch", Channel: strings.ToLower(targetChannel)}).Find(&channels)

	if len(channels) == 0 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Bot is not in channel %s", targetChannel),
			},
		}, nil
	}

	db.Delete(&channels)

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	go func() {
		time.Sleep(time.Millisecond * 500)
		if err := tw.Leave(targetChannel); err != nil {
			log.Printf("failed to leave channel %s: %v", targetChannel, err)
		}
	}()

	var msgs []*message.Message
	if strings.EqualFold(msg.Message.Channel, targetChannel) {
		msgs = append(msgs, &message.Message{
			Channel: msg.Message.Channel,
			Text:    "Successfully left channel.",
		})
	} else {
		msgs = append(msgs, &message.Message{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Successfully left channel %s", targetChannel),
		})
	}
	return msgs, nil
}

func setPrefix(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := setPrefixPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) < 2 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "No new prefix provided",
			},
		}, nil
	}
	newPrefix := matches[1]

	db := database.Instance

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	var channels []model.JoinedChannel
	db.Where(model.JoinedChannel{Platform: "Twitch", Channel: strings.ToLower(msg.Message.Channel)}).Find(&channels)

	for _, channel := range channels {
		channel.Prefix = newPrefix

		result := db.Save(&channel)

		if result.RowsAffected == 0 {
			log.Printf("Failed to update prefix: %v", result.Error)
			return []*message.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Failed to update prefix",
				},
			}, nil
		}
	}

	tw.UpdateCachedJoinedChannels()

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Prefix set to %s", newPrefix),
		},
	}, nil
}
