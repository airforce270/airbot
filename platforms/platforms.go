package platforms

import (
	"airbot/commands"
	"airbot/config"
	"airbot/logs"
	"airbot/message"
	"airbot/platforms/twitch"

	"gorm.io/gorm"
)

// Platform represents a connection to a given platform (i.e. Twitch, Discord)
type Platform interface {
	// Name returns the platform's name.
	Name() string
	// Username returns the bot's username within the platform.
	Username() string

	// Listen returns a channel that will provide incoming messages.
	Listen() chan message.IncomingMessage
	// Send sends a message.
	Send(m message.Message) error

	// Connect connects to the platform.
	Connect() error
	// Disconnect disconnects from the platform and should be called before exiting.
	Disconnect() error
}

// Build builds connections to enabled platforms based on the config.
func Build(cfg *config.Config, db *gorm.DB) ([]Platform, error) {
	var p []Platform
	if twc := cfg.Platforms.Twitch; twc.Enabled {
		logs.Printf("Building Twitch platform...")
		p = append(p, twitch.New(twc.Username, twc.Channels, twc.AccessToken, twc.IsVerifiedBot, db))
	}
	return p, nil
}

// StartHandling starts handling commands coming from the given platform.
// This function blocks and should be run within a goroutine.
func StartHandling(p Platform, db *gorm.DB, logIncoming, logOutgoing bool) {
	handler := commands.Handler{}
	c := p.Listen()

	for {
		msg := <-c
		go processMessage(&handler, db, p, msg, logIncoming, logOutgoing)
	}
}

// processMessage processes a single message and may send a message in response.
func processMessage(handler *commands.Handler, db *gorm.DB, p Platform, msg message.IncomingMessage, logIncoming, logOutgoing bool) {
	if logIncoming {
		logs.Printf("[%s<- %s/%s]: %s", p.Name(), msg.Message.Channel, p.Username(), msg.Message.Text)
	}

	outMsgs, err := handler.Handle(&msg)
	if err != nil {
		logs.Printf("Failed to handle message %v: %v", msg, err)
		return
	}
	if len(outMsgs) == 0 {
		return
	}

	for _, out := range outMsgs {
		if logOutgoing {
			logs.Printf("[%s-> %s/%s]: %s", p.Name(), out.Channel, p.Username(), out.Text)
		}

		if err := p.Send(*out); err != nil {
			logs.Printf("Failed to send message %v: %v", msg, err)
		}
	}
}
