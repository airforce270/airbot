// Package provides helpers for testing connections to the Twitch TMI API.
package twitchtmitest

const (
	FetchChattersManyChattersResp  = `{"_links":{},"chatter_count":15,"chatters":{"broadcaster":["airforce2700"],"vips":[],"moderators":["af2bot","streamelements","fossabot","ip0g"],"staff":[],"admins":[],"global_mods":[],"viewers":["augustcelery","bapplesas","dafke_","ellagarten","esattt","femboynv","givemeanonion","iizzybeth","iqkev","rockn__"]}}`
	FetchChattersSingleChatterResp = `{"_links":{},"chatter_count":1,"chatters":{"broadcaster":["user2"],"vips":[],"moderators":[],"staff":[],"admins":[],"global_mods":[],"viewers":[]}}`
)
