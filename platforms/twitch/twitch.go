// Package twitch handles Twitch-specific logic.
package twitch

import (
	"fmt"

	"airbot/config"
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
	channels []config.TwitchChannelConfig
	// accessToken is the OAuth token to use when connecting.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	accessToken string
	// i is the twitch IRC client.
	i *twitchirc.Client
}

func (t *Twitch) Name() string { return "Twitch" }

func (t *Twitch) Username() string { return t.username }

func (t *Twitch) Send(m message.Message) error {
	go t.i.Say(m.Channel, m.Text)
	return nil
}

func (t *Twitch) Listen() chan message.IncomingMessage {
	c := make(chan message.IncomingMessage)
	t.i.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		c <- message.IncomingMessage{
			Message: message.Message{
				Text:    msg.Message,
				Channel: msg.Channel,
				User:    msg.User.Name,
				Time:    msg.Time,
			},
			Prefix: t.prefix(msg.Channel),
		}
	})
	return c
}

func (t *Twitch) Connect() error {
	logs.Printf("Creating Twitch IRC client...")
	i := twitchirc.NewClient(t.username, fmt.Sprintf("oauth:%s", t.accessToken))
	t.i = i

	if t.isVerifiedBot {
		t.i.SetJoinRateLimiter(twitchirc.CreateVerifiedRateLimiter())
	}

	for _, channel := range t.channels {
		logs.Printf("Joining Twitch channel %s...", channel.Name)
		i.Join(channel.Name)
	}

	logs.Printf("Connecting to Twitch IRC...")
	go func() {
		if err := t.i.Connect(); err != nil {
			logs.Printf("failed to connect to twitch IRC: %v", err)
		}
	}()

	return nil
}

func (t *Twitch) Disconnect() error {
	for _, channel := range t.channels {
		logs.Printf("Leaving Twitch channel %s...", channel.Name)
		t.i.Depart(channel.Name)
	}
	logs.Printf("Disconnecting from Twitch IRC...")
	return t.i.Disconnect()
}

func (t *Twitch) prefix(channel string) string {
	for _, ch := range t.channels {
		if ch.Name == channel {
			return ch.Prefix
		}
	}
	return ""
}

// New creates a new Twitch connection.
func New(username string, channels []config.TwitchChannelConfig, accessToken string, isVerifiedBot bool) *Twitch {
	return &Twitch{
		username:      username,
		channels:      channels,
		accessToken:   accessToken,
		isVerifiedBot: isVerifiedBot,
	}
}
