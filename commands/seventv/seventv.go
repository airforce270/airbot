// Package seventv implements 7TV commands.
package seventv

import (
	"context"
	"errors"
	"fmt"
	"log"

	seventvapi "github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	addEmoteCommand,
	emoteCountCommand,
	removeEmoteCommand,
}

var (
	addEmoteCommand = basecommand.Command{
		Name: "7tv add",
		Desc: "Adds an emote to a channel. Currently, the emote ID must be provided.",
		Params: []arg.Param{
			{Name: "emote id", Type: arg.String, Required: true},
			{Name: "alias", Type: arg.String, Required: false},
		},
		Permission: permission.Admin,
		Handler:    addEmote,
	}
	emoteCountCommand = basecommand.Command{
		Name:       "7tv emotecount",
		Desc:       "Counts emotes in a channel.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    emoteCount,
	}
	removeEmoteCommand = basecommand.Command{
		Name:       "7tv remove",
		Desc:       "Removes an emote from a channel. Currently, the emote ID must be provided.",
		Params:     []arg.Param{{Name: "emote id", Type: arg.String, Required: true}},
		Permission: permission.Admin,
		Handler:    removeEmote,
	}
)

func addEmote(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	emoteIDArg, aliasArg := args[0], args[1]
	if !emoteIDArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	emoteID := emoteIDArg.StringValue

	channel, err := msg.Resources.Platform.User(ctx, msg.Message.Channel)
	if err != nil {
		log.Printf("Looking up Twitch channel %s failed: %v", msg.Message.Channel, err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up Twitch channel %s failed (???)", msg.Message.Channel),
			},
		}, nil
	}
	channelTwitchID := channel.TwitchID
	if channelTwitchID == nil {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up Twitch channel %s failed - channel twitch ID unknown", msg.Message.Channel),
			},
		}, nil
	}
	userConnection, err := msg.Resources.Clients.SevenTV.FetchUserConnectionByTwitchUserId(*channelTwitchID)
	if err != nil {
		log.Printf("Failed to fetch 7TV user connection: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up 7TV user info for %s failed", msg.Message.Channel),
			},
		}, nil
	}
	emoteSetID := userConnection.EmoteSet.ID

	emoteName := emoteID

	if aliasArg.Present {
		alias := aliasArg.StringValue
		err = msg.Resources.Clients.SevenTV.AddEmoteWithAlias(ctx, emoteSetID, emoteID, alias)
		emoteName = alias
	} else {
		err = msg.Resources.Clients.SevenTV.AddEmote(ctx, emoteSetID, emoteID)
	}

	if err != nil {
		if errors.Is(err, seventvapi.ErrNotAuthorized) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Please add me as a 7TV editor if you'd like me to update emotes :)",
				},
			}, nil
		}
		if errors.Is(err, seventvapi.ErrEmoteAlreadyEnabled) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Emote is already enabled",
				},
			}, nil
		}
		if errors.Is(err, seventvapi.ErrEmoteNotFound) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Emote not found",
				},
			}, nil
		}

		log.Printf("Failed to add 7TV emote: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Failed to add 7TV emote %s to %s/%s", emoteID, msg.Resources.Platform.Name(), msg.Message.Channel),
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Added 7TV emote %s to %s/%s", emoteName, msg.Resources.Platform.Name(), msg.Message.Channel),
		},
	}, nil
}

func emoteCount(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)

	plat, ok := msg.Resources.PlatformByName(twitch.Name)
	if !ok {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Twitch connection not configured",
			},
		}, nil
	}
	tw := plat.(*twitch.Twitch)

	user, err := tw.FetchUser(target)
	if err != nil {
		if errors.Is(err, twitch.ErrChannelNotFound) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("Channel %s not found", target),
				},
			}, nil
		}
		return nil, err
	}

	resp, err := msg.Resources.Clients.SevenTV.FetchUserConnectionByTwitchUserId(user.ID)
	if err != nil {
		log.Printf("Failed to fetch 7TV user connection: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up %s on 7TV failed", user.DisplayName),
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s has %d emotes on 7TV", target, len(resp.EmoteSet.Emotes)),
		},
	}, nil
}

func removeEmote(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	emoteIDArg := args[0]
	if !emoteIDArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	emoteID := emoteIDArg.StringValue

	channel, err := msg.Resources.Platform.User(ctx, msg.Message.Channel)
	if err != nil {
		log.Printf("Looking up Twitch channel %s failed: %v", msg.Message.Channel, err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up Twitch channel %s failed (???)", msg.Message.Channel),
			},
		}, nil
	}
	channelTwitchID := channel.TwitchID
	if channelTwitchID == nil {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up Twitch channel %s failed - channel twitch ID unknown", msg.Message.Channel),
			},
		}, nil
	}
	userConnection, err := msg.Resources.Clients.SevenTV.FetchUserConnectionByTwitchUserId(*channelTwitchID)
	if err != nil {
		log.Printf("Failed to fetch 7TV user connection: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up 7TV user info for %s failed", msg.Message.Channel),
			},
		}, nil
	}
	emoteSetID := userConnection.EmoteSet.ID

	err = msg.Resources.Clients.SevenTV.RemoveEmote(ctx, emoteSetID, emoteID)

	if err != nil {
		if errors.Is(err, seventvapi.ErrNotAuthorized) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Please add me as a 7TV editor if you'd like me to update emotes :)",
				},
			}, nil
		}
		if errors.Is(err, seventvapi.ErrEmoteNotEnabled) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Emote is not enabled",
				},
			}, nil
		}
		if errors.Is(err, seventvapi.ErrEmoteNotFound) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Emote not found",
				},
			}, nil
		}

		log.Printf("Failed to remove 7TV emote: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Failed to remove 7TV emote %s from %s/%s", emoteID, msg.Resources.Platform.Name(), msg.Message.Channel),
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Removed 7TV emote %s from %s/%s", emoteID, msg.Resources.Platform.Name(), msg.Message.Channel),
		},
	}, nil
}
