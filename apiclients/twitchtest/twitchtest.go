// Package provides helpers for testing connections to the Twitch API.
package twitchtest

var (
	GetChannelInformationResp  = `{"data":[{"broadcaster_id":"141981764","broadcaster_login":"user1","broadcaster_name":"user1","broadcaster_language":"en","game_id":"509670","game_name":"Science&Technology","title":"TwitchDevMonthlyUpdate//May6,2021","delay":0}]}`
	GetUsersResp               = `{"data":[{"id":"user2","login":"user2","display_name":"user2","type":"","broadcaster_type":"partner","description":"Supporting third-party developers building Twitch integrations from chatbots to game integrations.","profile_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/8a6381c7-d0c0-4576-b179-38bd5ce1d6af-profile_image-300x300.png","offline_image_url":"https://static-cdn.jtvnw.net/jtv_user_pictures/3f13ab61-ec78-4fe6-8481-8682cb3b0ac2-channel_offline_image-1920x1080.png","view_count":5980557,"email":"not-real@email.com","created_at":"2016-12-14T20:32:28Z"}]}`
	GetChannelChatChattersResp = `{"data":[{"user_id":"1","user_name":"user1","user_login":"user1"},{"user_id":"2","user_name":"user2","user_login":"user2"}]}`
)
