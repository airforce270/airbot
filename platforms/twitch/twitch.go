// Package twitch handles Twitch-specific logic.
package twitch

import (
	"fmt"

	"airbot/logs"
	"airbot/message"

	twitchirc "github.com/gempir/go-twitch-irc/v3"
)

// Twitch implements Platform for a connection to Twitch chat.
type Twitch struct {
	// username is the Twitch username the bot is running as.
	username string
	// isVerifiedBot is whether the user running as is a verified bot on Twitch.
	// See https://dev.twitch.tv/docs/irc#verified-bots
	isVerifiedBot bool
	// channels is the Twitch channels to join.
	channels []string
	// accessToken is the OAuth token to use when connecting.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	accessToken string
	// c is the channel to send incoming messages on.
	c chan message.Message
	// i is the twitch IRC connection.
	i *twitchirc.Client
}

func (t *Twitch) Name() string { return "Twitch" }

func (t *Twitch) Username() string { return t.username }

func (t *Twitch) Send(m message.Message) error {
	t.i.Say(m.Channel, m.Text)
	return nil
}

func (t *Twitch) Listen() chan message.Message {
	t.c = make(chan message.Message)
	t.i.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		t.c <- message.Message{
			Text:    msg.Message,
			Channel: msg.Channel,
			User:    msg.User.Name,
			Time:    msg.Time,
		}
	})
	return t.c
}

func (t *Twitch) Connect() error {
	logs.Printf("Creating Twitch IRC client...")
	i := twitchirc.NewClient(t.username, fmt.Sprintf("oauth:%s", t.accessToken))
	t.i = i

	if t.isVerifiedBot {
		t.i.SetJoinRateLimiter(twitchirc.CreateVerifiedRateLimiter())
	}

	for _, channel := range t.channels {
		logs.Printf("Joining Twitch channel %s...", channel)
		i.Join(channel)
	}

	logs.Printf("Connecting to Twitch IRC...")
	go (func() error {
		if err := t.i.Connect(); err != nil {
			return fmt.Errorf("failed to connect to twitch IRC: %w", err)
		}
		return nil
	})()

	return nil
}

func (t *Twitch) Disconnect() error {
	for _, channel := range t.channels {
		logs.Printf("Leaving Twitch channel %s...", channel)
		t.i.Depart(channel)
	}
	logs.Printf("Disconnecting from Twitch IRC...")
	return t.i.Disconnect()
}

// New creates a new Twitch.
func New(username string, channels []string, accessToken string, isVerifiedBot bool) *Twitch {
	return &Twitch{
		username:      username,
		channels:      channels,
		accessToken:   accessToken,
		isVerifiedBot: isVerifiedBot,
	}
}
