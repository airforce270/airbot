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
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/model"
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
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
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

const defaultPrefix = "$"

func joinChannel(msg *base.IncomingMessage, targetChannel string) ([]*base.Message, error) {
	prefix := defaultPrefix

	db := database.Instance

	var channels []model.JoinedChannel
	db.Where(model.JoinedChannel{
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

	channelRecord := model.JoinedChannel{
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

	var channels []model.JoinedChannel
	db.Where(model.JoinedChannel{Platform: msg.Platform.Name(), Channel: strings.ToLower(msg.Message.Channel)}).Find(&channels)

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
