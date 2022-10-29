// Package config handles reading the config data.
package config

import (
	"encoding/json"
	"os"
)

// Path contains the path to the config file to be read (when referenced by the binary).
const Path = "config.json"

// Config is the top-level config object.
type Config struct {
	// LogIncoming is whether the bot should log incoming messages.
	LogIncoming bool `json:"logIncomingMessages"`
	// LogOutgoing is whether the bot should log outgoing messages.
	LogOutgoing bool `json:"logOutgoingMessages"`
	// Platforms contains platform-specific config data.
	Platforms PlatformConfig `json:"platforms"`
	// EnableNonPrefixCommands is whether non-prefix commands should be enabled.
	EnableNonPrefixCommands bool `json:"enableNonPrefixCommands"`
	// Supinic contains config for talking to the Supinic API.
	Supinic SupinicConfig `json:"supinic"`
}

// PlatformConfig is platform-specific config data.
type PlatformConfig struct {
	// Twitch contains Twitch-specific config data.
	Twitch TwitchConfig `json:"twitch"`
}

// TwitchConfig is Twitch-specific config data.
type TwitchConfig struct {
	// Enabled is whether Twitch should be connected to and messages handled.
	Enabled bool `json:"enabled"`
	// Username is the Twitch username of the account to use for the bot.
	Username string `json:"username"`
	// ClientID is the Twitch Client ID of the bot.
	ClientID string `json:"clientId"`
	// AccessToken is a **user** OAuth2 token generated by Twitch for the bot account.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	AccessToken string `json:"accessToken"`
	// Owners contains the Twitch usernames of the bot owner(s).
	Owners []string `json:"owners"`
}

const (
	// placeholderSupinicUserID is the placeholder UserID for the Supinic API.
	// If the user ID is this value, it means the user hasn't configured it.
	placeholderSupinicUserID = "not-required-to-run-bot"
	// placeholderSupinicAPIKey is the placeholder APIKey for the Supinic API.
	// If the API key is this value, it means the user hasn't configured it.
	placeholderSupinicAPIKey = "you-can-safely-leave-this-as-is"
)

// SupinicConfig contains data for talking to the Supinic API.
type SupinicConfig struct {
	// UserID is the Supinic User ID of the bot.
	// https://supinic.com/user/auth-key
	UserID string `json:"userId"`
	// APIKey is an authentication key for the bot.
	// https://supinic.com/user/auth-key
	APIKey string `json:"apiKey"`
	// ShouldPingAPI is whether the Supinic API should be pinged from time to time
	// to let the Supinic API know that the bot is alive.
	// Normally, this should only be done by af2bot, the reference instance of Airbot.
	ShouldPingAPI bool `json:"shouldPingApi"`
}

func (s *SupinicConfig) IsConfigured() bool {
	hasDefaultValue := s.UserID == placeholderSupinicUserID || s.APIKey == placeholderSupinicAPIKey
	isUnset := s.UserID == "" || s.APIKey == ""
	return !hasDefaultValue && !isUnset
}

// Read reads the config data from the given path.
func Read(path string) (*Config, error) {
	raw, err := OSReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(raw)
}

var (
	OSReadFile = os.ReadFile
)

// parse parses raw bytes into a config.
func parse(data []byte) (*Config, error) {
	cfg := Config{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
