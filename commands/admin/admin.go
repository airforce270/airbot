// Package admin handles bot administration commands.
package admin

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/message"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"

	"golang.org/x/exp/slices"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	joinCommand,
	joinOtherCommand,
	leaveCommand,
	leaveOtherCommand,
}

var (
	joinCommandPattern = basecommand.PrefixPattern("join$")
	joinCommand        = basecommand.Command{
		Name:    "join",
		Help:    "Tells the bot to join your chat.",
		Pattern: joinCommandPattern,
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
			return joinChannel(msg, msg.Message.User)
		},
		PrefixOnly: true,
	}

	joinOtherCommandPattern = basecommand.PrefixPattern("joinother")
	joinOtherCommand        = basecommand.Command{
		Name:    "joinother",
		Help:    "Tells the bot to join a chat.",
		Pattern: joinOtherCommandPattern,
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
			return joinChannel(msg, basecommand.ParseTarget(msg, joinOtherPattern))
		},
		PrefixOnly: true,
		AdminOnly:  true,
	}
	joinOtherPattern = regexp.MustCompile(joinOtherCommandPattern.String() + `(\w+).*`)

	leaveCommandPattern = basecommand.PrefixPattern("leave$")
	leaveCommand        = basecommand.Command{
		Name:    "leave",
		Help:    "Tells the bot to leave your chat.",
		Pattern: leaveCommandPattern,
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
			return leaveChannel(msg, msg.Message.User)
		},
		PrefixOnly: true,
	}

	leaveOtherCommandPattern = basecommand.PrefixPattern("leaveother")
	leaveOtherCommand        = basecommand.Command{
		Name:    "leaveother",
		Help:    "Tells the bot to leave a chat.",
		Pattern: leaveOtherCommandPattern,
		Handler: func(msg *message.IncomingMessage) ([]*message.Message, error) {
			return leaveChannel(msg, basecommand.ParseTarget(msg, leaveOtherPattern))
		},
		PrefixOnly: true,
		AdminOnly:  true,
	}
	leaveOtherPattern = regexp.MustCompile(leaveOtherCommandPattern.String() + `(\w+).*`)
)

const defaultPrefix = "$"

func joinChannel(msg *message.IncomingMessage, targetChannel string) ([]*message.Message, error) {
	prefix := defaultPrefix

	cfg := config.Instance
	if cfg == nil {
		return nil, fmt.Errorf("config not initialized (was it set in main after reading?)")
	}

	for _, alreadyJoinedChannel := range cfg.Platforms.Twitch.Channels {
		if strings.EqualFold(alreadyJoinedChannel.Name, targetChannel) {
			return []*message.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("Channel %s is already joined", targetChannel),
				},
			}, nil
		}
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
		return nil, fmt.Errorf("failed to write new config: %w", err)
	}

	channelConfig := config.TwitchChannelConfig{
		Name:   channel.BroadcasterName,
		Prefix: prefix,
	}
	cfg.Platforms.Twitch.Channels = append(cfg.Platforms.Twitch.Channels, channelConfig)
	if err := config.Write(config.Path, cfg); err != nil {
		return nil, fmt.Errorf("failed to write new config: %w", err)
	}

	if err := tw.Join(channel, channelConfig); err != nil {
		return nil, fmt.Errorf("failed to join channel %s: %w", channel.BroadcasterName, err)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Successfully joined channel %s with prefix %s", channel.BroadcasterName, prefix),
		},
	}, nil
}

func leaveChannel(msg *message.IncomingMessage, targetChannel string) ([]*message.Message, error) {
	cfg := config.Instance
	if cfg == nil {
		return nil, fmt.Errorf("config not initialized (was it set in main after reading?)")
	}

	var alreadyJoinedChannels []string
	for _, alreadyJoinedChannel := range cfg.Platforms.Twitch.Channels {
		alreadyJoinedChannels = append(alreadyJoinedChannels, strings.ToLower(alreadyJoinedChannel.Name))
	}

	if !slices.Contains(alreadyJoinedChannels, strings.ToLower(targetChannel)) {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Bot is not in channel %s", targetChannel),
			},
		}, nil
	}

	var newChannelConfigs []config.TwitchChannelConfig
	for _, channelConfig := range cfg.Platforms.Twitch.Channels {
		if strings.EqualFold(channelConfig.Name, targetChannel) {
			continue
		}
		newChannelConfigs = append(newChannelConfigs, channelConfig)
	}
	cfg.Platforms.Twitch.Channels = newChannelConfigs
	if err := config.Write(config.Path, cfg); err != nil {
		return nil, fmt.Errorf("failed to write new config: %w", err)
	}

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	defer func() {
		if err := tw.Leave(targetChannel); err != nil {
			log.Printf("failed to leave channel %s: %v", targetChannel, err)
		}
	}()

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Successfully left channel %s", targetChannel),
		},
	}, nil
}
