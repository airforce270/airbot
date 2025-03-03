// Package restart provides a way for a command to restart the bot.
package restart

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"golang.org/x/sync/errgroup"
)

// C should be sent a value when the bot should be restarted.
var C = make(chan bool, 1)

// WriteRequester writes information about where the restart was requested from.
func WriteRequester(ctx context.Context, c cache.Cache, platform, channel, id string) {
	const expireIn = 30 * time.Second

	go func() {
		if err := c.StoreExpiringString(ctx, cache.KeyRestartRequestedOnPlatform, platform, expireIn); err != nil {
			log.Printf("Failed to store platform that restart was requested from (%s): %v", platform, err)
		}
	}()
	go func() {
		if err := c.StoreExpiringString(ctx, cache.KeyRestartRequestedInChannel, channel, expireIn); err != nil {
			log.Printf("Failed to store channel that restart was requested from (%s): %v", channel, err)
		}
	}()
	go func() {
		if err := c.StoreExpiringString(ctx, cache.KeyRestartRequestedByMessageID, id, expireIn); err != nil {
			log.Printf("Failed to store message ID that requested restart (%s): %v", id, err)
		}
	}()
}

// Notify notifies interested parties that the restart has finished.
func Notify(ctx context.Context, c cache.Cache, platforms map[string]base.Platform) error {
	platformCh := make(chan string, 1)
	channelCh := make(chan string, 1)
	messageCh := make(chan string, 1)

	var eg errgroup.Group

	eg.Go(func() error {
		platform, err := c.FetchString(ctx, cache.KeyRestartRequestedOnPlatform)
		if err != nil {
			return fmt.Errorf("failed to fetch platform that restart was requested from: %w", err)
		}
		platformCh <- platform
		return nil
	})
	eg.Go(func() error {
		channel, err := c.FetchString(ctx, cache.KeyRestartRequestedInChannel)
		if err != nil {
			return fmt.Errorf("failed to fetch channel that restart was requested from: %w", err)
		}
		channelCh <- channel
		return nil
	})
	eg.Go(func() error {
		messageID, err := c.FetchString(ctx, cache.KeyRestartRequestedByMessageID)
		if err != nil {
			return fmt.Errorf("failed to fetch message that requested restart: %w", err)
		}
		messageCh <- messageID
		return nil
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to retrieve values: %w", err)
	}

	platform := <-platformCh
	channel := <-channelCh
	messageID := <-messageCh

	for pName, p := range platforms {
		if pName != platform {
			continue
		}
		msg := base.Message{
			Text:    "Airbot has restarted.",
			Channel: channel,
		}
		if messageID != "" {
			if err := p.Reply(ctx, msg, messageID); err != nil {
				return fmt.Errorf("failed to notify restart to %s/%s: %w", platform, channel, err)
			}
		} else {
			if err := p.Send(ctx, msg); err != nil {
				return fmt.Errorf("failed to notify restart to %s/%s: %w", platform, channel, err)
			}
		}
	}

	return nil
}
