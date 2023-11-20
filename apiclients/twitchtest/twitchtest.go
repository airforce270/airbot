// Package provides helpers for testing connections to the Twitch API.
package twitchtest

import _ "embed"

var (
	//go:embed get_channel_information.json
	GetChannelInformationResp string
	//go:embed get_users.json
	GetUsersResp string
	//go:embed get_channel_chat_chatters.json
	GetChannelChatChattersResp string
)
