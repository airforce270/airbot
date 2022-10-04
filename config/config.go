// Package config handles reading the config data.
package config

import (
	"encoding/json"
	"os"
)

// Config is the top-level config object.
type Config struct {
	// LogIncoming is whether the bot should log incoming messages.
	LogIncoming bool `json:"logIncomingMessages"`
	// LogOutgoing is whether the bot should log outgoing messages.
	LogOutgoing bool `json:"logOutgoingMessages"`
	// Platforms contains platform-specific config data.
	Platforms platformConfig `json:"platforms"`
}

// platformConfig is platform-specific config data.
type platformConfig struct {
	// Twitch contains Twitch-specific config data.
	Twitch twitchConfig `json:"twitch"`
}

// twitchConfig is Twitch-specific config data.
type twitchConfig struct {
	// Enabled is whether Twitch should be connected to and messages handled.
	Enabled bool `json:"enabled"`
	// AccessToken is a **user** OAuth2 token generated by Twitch.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	AccessToken string `json:"accessToken"`
	// Username is the Twitch username of the account to use for the bot.
	Username string `json:"username"`
	// IsVerifiedBot is whether the bot's account is a verified bot on Twitch.
	// See https://dev.twitch.tv/docs/irc#verified-bots
	IsVerifiedBot bool `json:"isVerifiedBot"`
	// Twitch channels the bot should join and listen to messages in.
	// Should be channel names, not IDs.
	Channels []string `json:"channels"`
}

// Read reads the config data from the given path.
func Read(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(raw)
}

// parse parses raw bytes into a config.
func parse(data []byte) (*Config, error) {
	cfg := Config{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
