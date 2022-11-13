// Package platforms contains the Platform interface and related.
package platforms

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/commands"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/platforms/twitch"

	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

// Build builds connections to enabled platforms based on the config.
func Build(cfg *config.Config, db *gorm.DB, cdb *redis.Client) (map[string]base.Platform, error) {
	p := map[string]base.Platform{}
	if twc := cfg.Platforms.Twitch; twc.Enabled {
		log.Printf("Building Twitch platform...")
		tw := twitch.New(twc.Username, twc.Owners, twc.ClientID, twc.AccessToken, db, cdb)
		twitch.Instance = tw
		p[tw.Name()] = tw
	}
	return p, nil
}

// StartHandling starts handling commands coming from the given platform.
// This function blocks and should be run within a goroutine.
func StartHandling(p base.Platform, db *gorm.DB, cdb *redis.Client, logIncoming, logOutgoing, enableNonPrefixCommands bool) {
	handler := commands.NewHandler(db, enableNonPrefixCommands)
	inC := p.Listen()

	outC := make(chan base.Message)
	go startSending(p, outC, cdb, logOutgoing)

	for {
		msg := <-inC
		go processMessage(&handler, db, p, outC, msg, logIncoming)
	}
}

// processMessage processes a single message and may queue messages to be sent in response.
func processMessage(handler *commands.Handler, db *gorm.DB, p base.Platform, outC chan<- base.Message, msg base.IncomingMessage, logIncoming bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("processMessage panicked, recovered: %v; %s", r, debug.Stack())
		}
	}()

	if logIncoming {
		log.Printf("[%s<- %s/%s]: %s", p.Name(), msg.Message.Channel, msg.Message.User, msg.Message.Text)
	}

	outMsgs, err := handler.Handle(&msg)
	if err != nil {
		log.Printf("Failed to handle message %v: %v", msg, err)
		return
	}
	if len(outMsgs) == 0 {
		return
	}

	for _, out := range outMsgs {
		outC <- *out
	}
}

// startSending sends messages from the queue.
func startSending(p base.Platform, outC <-chan base.Message, cdb *redis.Client, logOutgoing bool) {
	for {
		out := <-outC

		if logOutgoing {
			log.Printf("[%s-> %s/%s]: %s", p.Name(), out.Channel, p.Username(), out.Text)
		}

		if err := p.Send(out); err != nil {
			log.Printf("Failed to send message %v: %v", out, err)
		}

		slowmode, err := cache.FetchSlowmode(p, cdb)
		if err != nil {
			log.Printf("Failed to fetch slowmode status for %s: %v", p.Name(), err)
		}

		if slowmode {
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}
