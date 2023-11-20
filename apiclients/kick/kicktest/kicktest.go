package kicktest

import _ "embed"

var (
	//go:embed get_channel/large_live.json
	LargeLiveGetChannelResp string
	//go:embed get_channel/large_offline.json
	LargeOfflineGetChannelResp string
	//go:embed get_channel/small_live.json
	SmallLiveGetChannelResp string
	//go:embed get_channel/small_offline.json
	SmallOfflineGetChannelResp string
)
