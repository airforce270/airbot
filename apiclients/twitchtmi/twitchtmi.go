// Package twitchtmi provides an API client for the Twitch TMI API.
package twitchtmi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Base URL for API requests. Should only be changed for testing.
var BaseURL = "https://tmi.twitch.tv"

// FetchChattersResponse contains the response from the TMI API for chatters in a chat.
type FetchChattersResponse struct {
	// ChatterCount is the number of chatters in chat.
	ChatterCount int `json:"chatter_count"`
	// Chatters contains the organized chatters in chat.
	Chatters Chatters `json:"chatters"`
}

func (r FetchChattersResponse) AllChatters() []string {
	var chatters []string
	chatters = append(chatters, r.Chatters.Broadcaster...)
	chatters = append(chatters, r.Chatters.VIPs...)
	chatters = append(chatters, r.Chatters.Moderators...)
	chatters = append(chatters, r.Chatters.Staff...)
	chatters = append(chatters, r.Chatters.Admins...)
	chatters = append(chatters, r.Chatters.GlobalMods...)
	chatters = append(chatters, r.Chatters.Viewers...)
	return chatters
}

// Chatters contains information about the chatters currently in chat.
type Chatters struct {
	// Broadcaster is the broadcaster, if they are currently in chat.
	// Length will be either 0 or 1.
	Broadcaster []string `json:"broadcaster"`
	// VIPs are the VIPs currently in chat.
	VIPs []string `json:"vips"`
	// Moderators are the moderators currently in chat.
	Moderators []string `json:"moderators"`
	// Staff are the Twitch staff members currently in chat.
	Staff []string `json:"staff"`
	// Admins are the Twitch admins currently in chat.
	Admins []string `json:"admins"`
	// GlobalMods are the Twitch global mods currently in chat.
	GlobalMods []string `json:"global_mods"`
	// Viewers are the normal chatters currently in chat.
	Viewers []string `json:"viewers"`
}

// FetchChatters fetches the current chatters in a Twitch chat.
func FetchChatters(channel string) (*FetchChattersResponse, error) {
	body, err := get(fmt.Sprintf("%s/group/user/%s/chatters", BaseURL, strings.ToLower(channel)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chatters for %s: %w", channel, err)
	}

	resp := FetchChattersResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from Twitch TMI API: %w", err)
	}
	return &resp, nil
}

func get(reqURL string) (respBody []byte, err error) {
	httpResp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from Twitch TMI API (URL:%s): %v", reqURL, httpResp)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from Twitch TMI API: %w", err)
	}

	return body, nil
}
