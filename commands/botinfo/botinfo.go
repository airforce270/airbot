// Package botinfo implements commands that return info about the bot.
package botinfo

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/model"
	"github.com/airforce270/airbot/permission"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Name:           "botinfo",
		AlternateNames: []string{"bot", "info"},
		Help:           "Replies with info about the bot.",
		Usage:          "$botinfo",
		Permission:     permission.Normal,
		PrefixOnly:     false,
		Pattern:        basecommand.PrefixPattern("bot|info|botinfo"),
		Handler:        botinfo,
	},
	{
		Name:       "prefix",
		Help:       "Replies with the prefix in this channel.",
		Usage:      "$prefix",
		Permission: permission.Normal,
		PrefixOnly: false,
		Pattern:    regexp.MustCompile(`\s*(^|(wh?at( i|')?s (the |air|af2)(bot('?s)?)? ?))prefix\??\s*`),
		Handler:    prefix,
	},
	{
		Name:       "source",
		Help:       "Replies a link to the bot's source code.",
		Usage:      "$source",
		Permission: permission.Normal,
		PrefixOnly: false,
		Pattern:    basecommand.PrefixPattern("source"),
		Handler:    source,
	},
	{
		Name:       "stats",
		Help:       "Replies with stats about the bot.",
		Usage:      "$stats",
		Permission: permission.Normal,
		PrefixOnly: false,
		Pattern:    basecommand.PrefixPattern("stats"),
		Handler:    stats,
	},
}

func botinfo(msg *base.IncomingMessage) ([]*base.Message, error) {
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Beep boop, this is Airbot running as %s in %s with prefix %s on %s. Made by airforce2700, source available on GitHub ( %ssource )", msg.Platform.Username(), msg.Message.Channel, msg.Prefix, msg.Platform.Name(), msg.Prefix),
		},
	}, nil
}

func prefix(msg *base.IncomingMessage) ([]*base.Message, error) {
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("This channel's prefix is %s", msg.Prefix),
		},
	}, nil
}

func source(msg *base.IncomingMessage) ([]*base.Message, error) {
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Source code for airbot available at https://github.com/airforce270/airbot",
		},
	}, nil
}

func stats(msg *base.IncomingMessage) ([]*base.Message, error) {
	cpuPercents, err := cpu.Percent(time.Millisecond*100, false)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve CPU percentage: %w", err)
	}
	cpuPercent := cpuPercents[0]

	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memory info: %w", err)
	}

	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve host info: %w", err)
	}

	osInfo := fmt.Sprintf("%s %s", sentenceCase(hostInfo.Platform), sentenceCase(hostInfo.OS))

	botProcess, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve host info: %w", err)
	}
	botProcessCreateTime, err := botProcess.CreateTime()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve host info: %w", err)
	}
	botUptime := time.Since(time.UnixMilli(botProcessCreateTime))

	systemUptime := time.Duration(hostInfo.Uptime) * time.Second

	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database instance not initialized")
	}
	var recentlyProcessedMessages int64
	db.Model(&model.Message{}).Where("created_at > ?", time.Now().Add(time.Second*-60)).Count(&recentlyProcessedMessages)
	var joinedChannels int64
	db.Model(&model.JoinedChannel{}).Count(&joinedChannels)

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Airbot running on %s, bot uptime: %s, system uptime: %s, CPU: %2.1f%%, RAM: %2.1f%%, processed %d messages in %d channels in the last 60 seconds", osInfo, botUptime, systemUptime, cpuPercent, memory.UsedPercent, recentlyProcessedMessages, joinedChannels),
		},
	}, nil
}

func sentenceCase(str string) string {
	if str == "" {
		return str
	}
	if len(str) == 1 {
		return strings.ToUpper(str)
	}
	return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
}
