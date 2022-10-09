// Package twitch implements Twitch commands.
package twitch

import (
	"airbot/commands/basecommand"
	"airbot/message"
	twitchplatform "airbot/platforms/twitch"
	"fmt"
	"regexp"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	currentGameCommand,
	titleCommand,
}

var (
	titleCommandPattern = basecommand.PrefixPattern("title")
	titleCommand        = basecommand.Command{
		Pattern:    titleCommandPattern,
		Handle:     title,
		PrefixOnly: true,
	}
	titlePattern = regexp.MustCompile(titleCommandPattern.String() + `(\w+).*`)

	currentGameCommandPattern = basecommand.PrefixPattern("currentgame")
	currentGameCommand        = basecommand.Command{
		Pattern:    currentGameCommandPattern,
		Handle:     currentGame,
		PrefixOnly: true,
	}
	currentGamePattern = regexp.MustCompile(currentGameCommandPattern.String() + `(\w+).*`)
)

func title(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := titlePattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := matches[1]

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	channel, err := tw.Channel(targetChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channel info for %s: %w", targetChannel, err)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's title: %s", channel.BroadcasterName, channel.Title),
		},
	}, nil
}

func currentGame(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := currentGamePattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := matches[1]

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	channel, err := tw.Channel(targetChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channel info for %s: %w", targetChannel, err)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s is currenly playing %s", channel.BroadcasterName, channel.GameName),
		},
	}, nil
}
