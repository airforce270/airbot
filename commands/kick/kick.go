// Package kick implements Kick commands.
package kick

import (
	"errors"
	"fmt"
	"strings"

	kickclient "github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	isLiveCommand,
	titleCommand,
}

var (
	isLiveCommand = basecommand.Command{
		Name:       "kickislive",
		Aliases:    []string{"kislive"},
		Desc:       "Replies with whether the Kick channel is currently live.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: true}},
		Permission: permission.Normal,
		Handler:    isLive,
	}
	titleCommand = basecommand.Command{
		Name:       "kicktitle",
		Aliases:    []string{"ktitle"},
		Desc:       "Replies with the title of the Kick channel. Currently only works if the channel is live.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: true}},
		Permission: permission.Normal,
		Handler:    title,
	}
)

func isLive(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	channelArg := args[0]
	if !channelArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	targetChannel := channelArg.StringValue

	channel, err := msg.Resources.Clients.Kick.FetchChannel(targetChannel)
	if err != nil {
		if errors.Is(err, kickclient.ErrChannelNotFound) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    targetChannel + " does not exist",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch Kick channel data for %s: %w", targetChannel, err)
	}

	var resp strings.Builder
	resp.WriteString(channel.Name)
	if channel.Livestream != nil {
		resp.WriteString(" is currently live on Kick, ")
		category := channel.Livestream.Categories[0]
		fmt.Fprintf(&resp, "streaming %s to %d viewers.", category.DisplayName, channel.Livestream.ViewerCount)
	} else {
		resp.WriteString(" is not currently live on Kick.")
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp.String(),
		},
	}, nil
}

func title(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	channelArg := args[0]
	if !channelArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	targetChannel := channelArg.StringValue

	channel, err := msg.Resources.Clients.Kick.FetchChannel(targetChannel)
	if err != nil {
		if errors.Is(err, kickclient.ErrChannelNotFound) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    targetChannel + " does not exist",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch Kick channel data for %s: %w", targetChannel, err)
	}

	if channel.Livestream == nil {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Currently Kick only returns the title for live channels, and %s is not currently live.", channel.Name),
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's title on Kick: %s", channel.Name, channel.Livestream.Title),
		},
	}, nil
}
