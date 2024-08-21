// Package botinfo implements commands that return info about the bot.
package botinfo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
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
		Aliases:    []string{"bot", "info", "about", "ping"},
		Desc:       "Replies with info about the bot.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			var resp strings.Builder
			fmt.Fprintf(&resp, "Beep boop, this is Airbot running as %s in %s", msg.Resources.Platform.Username(), msg.Message.Channel)
			fmt.Fprintf(&resp, " with prefix %s on %s.", msg.Prefix, msg.Resources.Platform.Name())
			fmt.Fprintf(&resp, " Made by airforce2700, source available on GitHub ( %ssource )", msg.Prefix)
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    resp.String(),
				},
			}, nil
		},
	},
	{
		Name:       "prefix",
		Desc:       "Replies with the prefix in this channel.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "This channel's prefix is " + msg.Prefix,
				},
			}, nil
		},
	},
	{
		Name:       "source",
		Desc:       "Replies a link to the bot's source code.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Source code for Airbot available at https://github.com/airforce270/airbot",
				},
			}, nil
		},
	},
	{
		Name:       "stats",
		Desc:       "Replies with stats about the bot.",
		Permission: permission.Normal,
		Handler:    stats,
	},
}

func stats(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	var g errgroup.Group

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

	const recentlyProcessedMessagesInterval = 60 * time.Second
	var recentlyProcessedMessages int64
	g.Go(func() error {
		err := msg.Resources.DB.Model(&models.Message{}).Where("created_at > ?", time.Now().Add(-recentlyProcessedMessagesInterval)).Count(&recentlyProcessedMessages).Error
		if err != nil {
			return fmt.Errorf("failed to count recently processed messages: %w", err)
		}
		return nil
	})

	var joinedChannels int64
	g.Go(func() error {
		err := msg.Resources.DB.Model(&models.JoinedChannel{}).Count(&joinedChannels).Error
		if err != nil {
			return fmt.Errorf("failed to count joined channels: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("submetric failed: %w", err)
	}

	var out strings.Builder
	fmt.Fprintf(&out, "Airbot running on %s %s", sentenceCase(hostInfo.Platform), sentenceCase(hostInfo.OS))
	if runningInDocker {
		out.WriteString(" (Docker)")
	}

	fmt.Fprintf(&out, ", bot uptime: %s", botUptime.Round(time.Second))
	fmt.Fprintf(&out, ", system uptime: %s", (time.Duration(hostInfo.Uptime) * time.Second).Round(time.Second))
	fmt.Fprintf(&out, ", CPU: %2.1f%%", cpuPercent)
	fmt.Fprintf(&out, ", RAM: %2.1f%%", memory.UsedPercent)
	fmt.Fprintf(&out, ", processed %d messages in %d channels in the last %d seconds", recentlyProcessedMessages, joinedChannels, int(recentlyProcessedMessagesInterval.Seconds()))

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    out.String(),
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
