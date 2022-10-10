// Package twitch implements Twitch commands.
package twitch

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	banReasonCommand,
	currentGameCommand,
	titleCommand,
	verifiedBotCommand,
}

var (
	banReasonCommandPattern = basecommand.PrefixPattern("(?:banreason|br)")
	banReasonCommand        = basecommand.Command{
		Pattern:    banReasonCommandPattern,
		Handle:     banReason,
		PrefixOnly: true,
	}
	banReasonPattern = regexp.MustCompile(banReasonCommandPattern.String() + `(\w+).*`)

	currentGameCommandPattern = basecommand.PrefixPattern("currentgame")
	currentGameCommand        = basecommand.Command{
		Pattern:    currentGameCommandPattern,
		Handle:     currentGame,
		PrefixOnly: true,
	}
	currentGamePattern = regexp.MustCompile(currentGameCommandPattern.String() + `(\w+).*`)

	titleCommandPattern = basecommand.PrefixPattern("title")
	titleCommand        = basecommand.Command{
		Pattern:    titleCommandPattern,
		Handle:     title,
		PrefixOnly: true,
	}
	titlePattern = regexp.MustCompile(titleCommandPattern.String() + `(\w+).*`)

	verifiedBotCommandPattern = basecommand.PrefixPattern("(?:verifiedbot|vb)")
	verifiedBotCommand        = basecommand.Command{
		Pattern:    verifiedBotCommandPattern,
		Handle:     verifiedBot,
		PrefixOnly: true,
	}
	verifiedBotPattern = regexp.MustCompile(verifiedBotCommandPattern.String() + `(\w+).*`)
)

func banReason(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := banReasonPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

	user, err := ivr.FetchUser(targetChannel)
	if err != nil {
		return nil, err
	}

	var resp string
	if !user.IsBanned {
		resp = fmt.Sprintf("%s is not banned.", user.DisplayName)
	} else {
		resp = fmt.Sprintf("%s's ban reason: %s", user.DisplayName, user.BanReason)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
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

func verifiedBot(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := verifiedBotPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

	user, err := ivr.FetchUser(targetChannel)
	if err != nil {
		return nil, err
	}

	var resp string
	if user.IsVerifiedBot {
		resp = fmt.Sprintf("%s is a verified bot. ✅", user.DisplayName)
	} else {
		resp = fmt.Sprintf("%s is not a verified bot. ❌", user.DisplayName)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}
