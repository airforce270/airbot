// Package botinfo implements commands that return info about the bot.
package botinfo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sync/errgroup"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Name:       "botinfo",
		Aliases:    []string{"bot", "info", "botinfo", "about", "ping"},
		Help:       "Replies with info about the bot.",
		Permission: permission.Normal,
		Handler:    botinfo,
	},
	{
		Name:       "prefix",
		Help:       "Replies with the prefix in this channel.",
		Permission: permission.Normal,
		Handler:    prefix,
	},
	{
		Name:       "source",
		Help:       "Replies a link to the bot's source code.",
		Permission: permission.Normal,
		Handler:    source,
	},
	{
		Name:       "stats",
		Help:       "Replies with stats about the bot.",
		Permission: permission.Normal,
		Handler:    stats,
	},
}

func botinfo(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Beep boop, this is Airbot running as %s in %s with prefix %s on %s. Made by airforce2700, source available on GitHub ( %ssource )", msg.Platform.Username(), msg.Message.Channel, msg.Prefix, msg.Platform.Name(), msg.Prefix),
		},
	}, nil
}

func prefix(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("This channel's prefix is %s", msg.Prefix),
		},
	}, nil
}

func source(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Source code for airbot available at https://github.com/airforce270/airbot",
		},
	}, nil
}

func stats(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database instance not initialized")
	}

	g := errgroup.Group{}

	var cpuPercent float64
	g.Go(func() error {
		cpuPercents, err := cpu.Percent(time.Millisecond*50, false)
		if err != nil {
			return fmt.Errorf("failed to retrieve CPU percentage: %w", err)
		}
		cpuPercent = cpuPercents[0]
		return nil
	})

	var memory *mem.VirtualMemoryStat
	g.Go(func() error {
		m, err := mem.VirtualMemory()
		if err != nil {
			return fmt.Errorf("failed to retrieve memory info: %w", err)
		}
		memory = m
		return nil
	})

	var hostInfo *host.InfoStat
	g.Go(func() error {
		hi, err := host.Info()
		if err != nil {
			return fmt.Errorf("failed to retrieve host info: %w", err)
		}
		hostInfo = hi
		return nil
	})

	var runningInDocker bool
	g.Go(func() error {
		// There's no reliable way to determine if we're running in Docker.
		// So we set this environment variable in our docker compose config.
		_, found := os.LookupEnv("RUNNING_IN_DOCKER")
		runningInDocker = found
		return nil
	})

	var botUptime time.Duration
	g.Go(func() error {
		botProcess, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			return fmt.Errorf("failed to find bot process: %w", err)
		}
		botProcessCreateTime, err := botProcess.CreateTime()
		if err != nil {
			return fmt.Errorf("failed to get bot startup time: %w", err)
		}
		botUptime = time.Since(time.UnixMilli(botProcessCreateTime))
		return nil
	})

	var recentlyProcessedMessages int64
	g.Go(func() error {
		db.Model(&models.Message{}).Where("created_at > ?", time.Now().Add(time.Second*-60)).Count(&recentlyProcessedMessages)
		return nil
	})

	var joinedChannels int64
	g.Go(func() error {
		db.Model(&models.JoinedChannel{}).Count(&joinedChannels)
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	startPart := fmt.Sprintf("Airbot running on %s %s", sentenceCase(hostInfo.Platform), sentenceCase(hostInfo.OS))
	if runningInDocker {
		startPart += " (Docker)"
	}

	parts := []string{
		startPart,
		fmt.Sprintf("bot uptime: %s", botUptime),
		fmt.Sprintf("system uptime: %s", time.Duration(hostInfo.Uptime)*time.Second),
		fmt.Sprintf("CPU: %2.1f%%", cpuPercent),
		fmt.Sprintf("RAM: %2.1f%%", memory.UsedPercent),
		fmt.Sprintf("processed %d messages in %d channels in the last 60 seconds", recentlyProcessedMessages, joinedChannels),
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    strings.Join(parts, ", "),
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
