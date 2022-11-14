[//]: # ( !!! DO NOT EDIT MANUALLY !!!  This is a generated file, any changes will be overwritten! )

<!-- markdownlint-disable line-length -->

# Commands

All commands assume the `$` prefix, but note that the prefix is configurable
per-channel (in `config.json`).
To find out what the prefix is in a channel, ask `what's airbot's prefix?`
in a chat.

Some commands include parameters.

If the parameter is wrapped in `<angle brackets>`, it's a **required** parameter.

If it's wrapped in `[square brackets]`, it's an **optional** parameter.

## Admin

### $botslowmode

- Sets the bot to follow a global (per-platform) 1 second slowmode.
- > Usage: `$botslowmode <on|off>`
- > Minimum permission level: `Owner`

### $echo

- Echoes back whatever is sent.
- > Usage: `$echo`
- > Minimum permission level: `Owner`

### $join

- Tells the bot to join your chat.
- > Usage: `$join`

### $joinother

- Tells the bot to join a chat.
- > Usage: `$joinother <channel>`
- > Minimum permission level: `Owner`

### $joined

- Lists the channels the bot is currently in.
- > Usage: `$joined`
- > Minimum permission level: `Owner`

### $leave

- Tells the bot to leave your chat.
- > Usage: `$leave`
- > Minimum permission level: `Admin`

### $leaveother

- Tells the bot to leave a chat.
- > Usage: `$leaveother <channel>`
- > Minimum permission level: `Owner`

### $setprefix

- Sets the bot's prefix in the channel.
- > Usage: `$setprefix`
- > Minimum permission level: `Admin`

## Bot info

### $help

- Displays help for a command.
- > Usage: `$help <command>`

### $botinfo

- Replies with info about the bot.
- > Usage: `$botinfo`
- > Alternate commands: `$bot`, `$info`, `$about`

### $prefix

- Replies with the prefix in this channel.
- > Usage: `$prefix`

### $source

- Replies a link to the bot's source code.
- > Usage: `$source`

### $stats

- Replies with stats about the bot.
- > Usage: `$stats`

## Bulk

### $filesay

- Runs all commands in a given pastebin file.
- > Usage: `$filesay <pastebin raw url>`
- > Minimum permission level: `Mod`

## Echo

### $commands

- Replies with a link to the commands.
- > Usage: `$commands`

### $gn

- Says good night.
- > Usage: `$gn`

### $pyramid

- Makes a pyramid in chat. Max width 25.
- > Usage: `$pyramid <width> <text>`
- > Minimum permission level: `Mod`
- > Per-channel cooldown: `30s`

### $spam

- Sends a message many times. Max amount 50.
- > Usage: `$spam <count> <text>`
- > Minimum permission level: `Mod`
- > Per-channel cooldown: `30s`

### $TriHard

- Replies with TriHard 7.
- > Usage: `$TriHard`

### $tuck

- Tuck someone to bed.
- > Usage: `$tuck <user>`

## Fun

### $bibleverse

- Looks up a bible verse.
- > Usage: `$bibleverse <book> <chapter:verse>`
- > Alternate commands: `$bv`

### $cock

- Tells you the length :)
- > Usage: `$cock [user]`

### $iq

- Tells you someone's IQ
- > Usage: `$iq [user]`

## Gamba

### $accept

- Accepts a duel.
- > Usage: `$accept`

### $decline

- Declines a duel.
- > Usage: `$decline`

### $duel

- Duels another chatter. They have 30 seconds to accept or decline.
- > Usage: `$points <user>`
- > Per-user cooldown: `5s`

### $points

- Checks how many points you have.
- > Usage: `$points [user]`
- > Alternate commands: `$p`

### $roulette

- Roulettes some points.
- > Usage: `$roulette <amount|percent%|all>`
- > Per-user cooldown: `5s`
- > Alternate commands: `$r`

## Moderation

### $vanish

- Times you out for 1 second.
- > Usage: `$vanish`

## Twitch

### $banreason

- Replies with the reason someone was banned on Twitch.
- > Usage: `$banreason <user>`
- > Alternate commands: `$br`

### $currentgame

- Replies with the game that's currently being streamed on a channel.
- > Usage: `$currentgame <channel>`

### $founders

- Replies with a channel's founders. If no channel is provided, the current channel will be used.
- > Usage: `$founders [channel]`

### $logs

- Replies with a link to a Twitch user's logs in a channel.
- > Usage: `$logs <channel> <user>`

### $mods

- Replies with a channel's mods. If no channel is provided, the current channel will be used.
- > Usage: `$mods [channel]`

### $namecolor

- Replies with a user's name color.
- > Usage: `$namecolor [user]`

### $title

- Replies with a channel's title. If no channel is provided, the current channel will be used.
- > Usage: `$title [channel]`

### $verifiedbot

- Replies whether a user is a verified bot.
- > Usage: `$verifiedbot [user]`
- > Alternate commands: `$vb`

### $verifiedbotquiet

- Replies whether a user is a verified bot, but responds quietly.
- > Usage: `$verifiedbotquiet [user]`
- > Alternate commands: `$vbq`

### $vips

- Replies with a channel's VIPs. If no channel is provided, the current channel will be used.
- > Usage: `$vips [channel]`
