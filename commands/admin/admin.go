// Package admin handles bot administration commands.
package admin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/permission"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/utils"
	"github.com/airforce270/airbot/utils/ptrs"
	"github.com/airforce270/airbot/utils/restart"
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
	reloadConfigCommand,
	restartCommand,
	setPrefixCommand,
}

const defaultPrefix = "$"

var (
	botSlowmodeCommand = basecommand.Command{
		Name: "botslowmode",
		Desc: "Sets the bot to follow a global (per-platform) 1 second slowmode. If no argument is provided, checks if slowmode is enabled.",
		Params: []arg.Param{
			{Name: "enable", Type: arg.Boolean, Required: false},
		},
		Permission: permission.Owner,
		Handler:    botSlowmode,
	}

	echoCommand = basecommand.Command{
		Name: "echo",
		Desc: "Echoes back whatever is sent.",
		Params: []arg.Param{
			{Name: "message", Type: arg.Variadic, Required: true},
		},
		Permission: permission.Owner,
		Handler: func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			valueArg := args[0]
			if !valueArg.Present {
				return nil, basecommand.ErrBadUsage
			}
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    valueArg.StringValue,
				},
			}, nil
		},
	}

	joinCommand = basecommand.Command{
		Name:       "join",
		Desc:       "Tells the bot to join your chat.",
		Params:     []arg.Param{{Name: "prefix", Type: arg.String, Required: false}},
		Permission: permission.Normal,
		Handler: func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			prefix := defaultPrefix
			if prefixArg := args[0]; prefixArg.Present {
				prefix = prefixArg.StringValue
			}
			return joinChannel(ctx, msg, msg.Message.User, prefix)
		},
	}

	joinedCommand = basecommand.Command{
		Name:       "joined",
		Desc:       "Lists the channels the bot is currently in.",
		Permission: permission.Owner,
		Handler:    joined,
	}

	joinOtherCommand = basecommand.Command{
		Name: "joinother",
		Desc: "Tells the bot to join a chat.",
		Params: []arg.Param{
			{Name: "channel", Type: arg.Username, Required: true},
			{Name: "prefix", Type: arg.String, Required: false},
		},
		Permission: permission.Owner,
		Handler: func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			channelArg := args[0]
			if !channelArg.Present {
				return nil, basecommand.ErrBadUsage
			}
			channel := channelArg.StringValue
			prefix := defaultPrefix
			if prefixArg := args[1]; prefixArg.Present {
				prefix = prefixArg.StringValue
			}
			return joinChannel(ctx, msg, channel, prefix)
		},
	}

	leaveCommand = basecommand.Command{
		Name:       "leave",
		Desc:       "Tells the bot to leave your chat.",
		Permission: permission.Admin,
		Handler: func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			return leaveChannel(ctx, msg, msg.Message.Channel)
		},
	}

	leaveOtherCommand = basecommand.Command{
		Name:       "leaveother",
		Desc:       "Tells the bot to leave a chat.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: true}},
		Permission: permission.Owner,
		Handler: func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			channelArg := args[0]
			if !channelArg.Present {
				return nil, basecommand.ErrBadUsage
			}
			return leaveChannel(ctx, msg, channelArg.StringValue)
		},
	}

	reloadConfigCommand = basecommand.Command{
		Name:       "reloadconfig",
		Desc:       "Reloads the bot's config after a config change.",
		Permission: permission.Admin,
		Handler:    reloadConfig,
	}

	restartCommand = basecommand.Command{
		Name:       "restart",
		Desc:       "Restarts the bot. Does not restart the database, etc.",
		Permission: permission.Admin,
		Handler:    restartBot,
	}

	setPrefixCommand = basecommand.Command{
		Name:       "setprefix",
		Desc:       "Sets the bot's prefix in the channel.",
		Params:     []arg.Param{{Name: "prefix", Type: arg.String, Required: true}},
		Permission: permission.Admin,
		Handler:    setPrefix,
	}
)

func botSlowmode(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	enableArg := args[0]
	key := cache.GlobalSlowmodeKey(msg.Resources.Platform.Name())

	if !enableArg.Present {
		enabled, err := msg.Resources.Cache.FetchBool(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch cache key %s (bool): %w", key, err)
		}
		enabledMsg := "enabled"
		if !enabled {
			enabledMsg = "disabled"
		}
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Bot slowmode is currently %s on %s", enabledMsg, msg.Resources.Platform.Name()),
			},
		}, nil
	}

	enable := enableArg.BoolValue

	if err := msg.Resources.Cache.StoreBool(ctx, key, enable); err != nil {
		log.Printf("Failed to set bot slowmode to %t on %s: %v", enable, msg.Resources.Platform.Name(), err)
		failureMsgStart := "Failed to enable"
		if !enable {
			failureMsgStart = "Failed to disable"
		}
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s bot slowmode on %s", failureMsgStart, msg.Resources.Platform.Name()),
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
			Text:    fmt.Sprintf("%s bot slowmode on %s", outMsgStart, msg.Resources.Platform.Name()),
		},
	}, nil
}

func joinChannel(ctx context.Context, msg *base.IncomingMessage, targetChannel, prefix string) ([]*base.Message, error) {
	channels, err := msg.Resources.Queries.SelectJoinedChannels(ctx, ptrs.StringNil(msg.Resources.Platform.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels matching %s/%s: %w", msg.Resources.Platform.Name(), strings.ToLower(targetChannel), err)
	}

	if len(channels) > 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Channel %s is already joined", targetChannel),
			},
		}, nil
	}

	_, err = msg.Resources.Queries.CreateJoinedChannel(ctx, database.CreateJoinedChannelParams{
		Platform: ptrs.StringNil(msg.Resources.Platform.Name()),
		Channel:  ptrs.StringNil(targetChannel),
		Prefix:   ptrs.StringNil(prefix),
		JoinedAt: ptrs.Ptr(time.Now()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to join channel %s: %w", targetChannel, err)
	}

	err = msg.Resources.Platform.Join(ctx, targetChannel, prefix)

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

func joined(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	joinedChannels, err := msg.Resources.Queries.SelectJoinedChannels(ctx, ptrs.StringNil(msg.Resources.Platform.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels matching %s: %w", msg.Resources.Platform.Name(), err)
	}
	var channels []string
	for _, c := range joinedChannels {
		channels = append(channels, *c.Channel)
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

func leaveChannel(ctx context.Context, msg *base.IncomingMessage, targetChannel string) ([]*base.Message, error) {
	err := database.LeaveChannel(ctx, msg.Resources.Queries, msg.Resources.Platform.Name(), targetChannel)

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
		if err := msg.Resources.Platform.Leave(targetChannel); err != nil {
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

func reloadConfig(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	configSrc, err := msg.Resources.NewConfigSource()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Read(configSrc)
	if err != nil {
		return nil, err
	}
	msg.Resources.Clients.Kick.Mtx.Lock()
	msg.Resources.Clients.Kick.JA3 = cfg.Platforms.Kick.JA3
	msg.Resources.Clients.Kick.UserAgent = cfg.Platforms.Kick.UserAgent
	msg.Resources.Clients.Kick.Mtx.Unlock()
	msg.Resources.Clients.SevenTV.SetAccessToken(cfg.SevenTV.AccessToken)

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Reloaded config.",
		},
	}, nil
}

func restartBot(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	go restart.WriteRequester(ctx, msg.Resources.Cache, msg.Resources.Platform.Name(), msg.Message.Channel, msg.Message.ID)

	const delay = 100 * time.Millisecond
	time.AfterFunc(delay, func() { restart.C <- true })

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Restarting Airbot.",
		},
	}, nil
}

func setPrefix(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	prefixArg := args[0]
	if !prefixArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	newPrefix := prefixArg.StringValue

	channel, err := msg.Resources.Queries.SelectJoinedChannel(ctx, database.SelectJoinedChannelParams{
		Platform: ptrs.StringNil(msg.Resources.Platform.Name()),
		Channel:  ptrs.StringNil(strings.ToLower(msg.Message.Channel)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel matching %s/%s: %w", msg.Resources.Platform.Name(), strings.ToLower(msg.Message.Channel), err)
	}

	affectedRows, err := msg.Resources.Queries.SetJoinedChannelPrefix(ctx, database.SetJoinedChannelPrefixParams{
		Platform: ptrs.StringNil(msg.Resources.Platform.Name()),
		Channel:  ptrs.StringNil(strings.ToLower(msg.Message.Channel)),
		Prefix:   &newPrefix,
	})
	if err != nil {
		log.Printf("Failed to save new prefix %s for channel %s: %v", newPrefix, *channel.Channel, err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Failed to update prefix",
			},
		}, nil
	}

	if affectedRows == 0 {
		log.Printf("Failed to update prefix, no rows affected")
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Failed to update prefix",
			},
		}, nil
	}

	if err := msg.Resources.Platform.SetPrefix(msg.Message.Channel, newPrefix); err != nil {
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
