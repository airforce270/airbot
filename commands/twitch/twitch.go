// Package twitch implements Twitch commands.
package twitch

import (
	"fmt"
	"regexp"
	"strings"

	"airbot/apiclients/ivr"
	"airbot/commands/basecommand"
	"airbot/message"
	twitchplatform "airbot/platforms/twitch"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	currentGameCommand,
	titleCommand,
	verifiedBotCommand,
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

	verifiedBotCommandPattern = basecommand.PrefixPattern("(?:vb|verifiedbot)")
	verifiedBotCommand        = basecommand.Command{
		Pattern:    verifiedBotCommandPattern,
		Handle:     verifiedBot,
		PrefixOnly: true,
	}
	verifiedBotPattern = regexp.MustCompile(verifiedBotCommandPattern.String() + `(\w+).*`)
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

func verifiedBot(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := verifiedBotPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

	isVerifiedBot, err := ivr.IsVerifiedBot(targetChannel)
	if err != nil {
		return nil, err
	}

	var resp string
	if isVerifiedBot {
		resp = fmt.Sprintf("%s is a verified bot. ✅", targetChannel)
	} else {
		resp = fmt.Sprintf("%s is not a verified bot. ❌", targetChannel)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}
