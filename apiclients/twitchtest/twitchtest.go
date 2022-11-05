// Package provides helpers for testing connections to the Twitch API.
package twitchtest

var (
	GetChannelInformationResp = `{"data":[{"broadcaster_id":"141981764","broadcaster_login":"user1","broadcaster_name":"user1","broadcaster_language":"en","game_id":"509670","game_name":"Science&Technology","title":"TwitchDevMonthlyUpdate//May6,2021","delay":0}]}`
	BanUserResp               = `{"data":[{"broadcaster_id":"1234","moderator_id":"5678","user_id":"9876","created_at":"2021-09-28T19:27:31Z","end_time":"2021-09-28T19:22:31Z"}]}`
)
