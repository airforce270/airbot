[//]: # ( !!! DO NOT EDIT MANUALLY !!!  This is a generated file, any changes will be overwritten! )

<!-- markdownlint-disable line-length -->

# Commands

All commands assume the `$` prefix, but the prefix is configurable
per-channel (see [$setprefix](#setprefix)).

Some commands include parameters.

If the parameter is wrapped in `<angle brackets>`, it's a **required** parameter.

If it's wrapped in `[square brackets]`, it's an **optional** parameter.

## 7TV

### $7tv add

- Adds an emote to a channel. Currently, the emote ID must be provided.
- > Usage: `$7tv add <emote id> [alias]`
- > Minimum permission level: `Admin`

### $7tv emotecount

- Counts emotes in a channel.
- > Usage: `$7tv emotecount [channel]`

### $7tv remove

- Removes an emote from a channel. Currently, the emote ID must be provided.
- > Usage: `$7tv remove <emote id>`
- > Minimum permission level: `Admin`

## Admin

### $botslowmode

- Sets the bot to follow a global (per-platform) 1 second slowmode. If no argument is provided, checks if slowmode is enabled.
- > Usage: `$botslowmode [on|off]`
- > Minimum permission level: `Owner`

### $echo

- Echoes back whatever is sent.
- > Usage: `$echo <message>`
- > Minimum permission level: `Owner`

### $join

- Tells the bot to join your chat.
- > Usage: `$join [prefix]`

### $joinother

- Tells the bot to join a chat.
- > Usage: `$joinother <channel> [prefix]`
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

### $reloadconfig

- Reloads the bot's config after a config change.
- > Usage: `$reloadconfig`
- > Minimum permission level: `Admin`

### $restart

- Restarts the bot. Does not restart the database, etc.
- > Usage: `$restart`
- > Minimum permission level: `Admin`

### $setprefix

- Sets the bot's prefix in the channel.
- > Usage: `$setprefix <prefix>`
- > Minimum permission level: `Admin`

## Bot info

### $help

- Displays help for a command.
- > Usage: `$help [command]`

### $botinfo

- Replies with info about the bot.
- > Usage: `$botinfo`
- > Aliases: `$bot`, `$info`, `$about`, `$ping`

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
- > Usage: `$filesay <pastebin raw URL>`
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

### $trihard

- Replies with TriHard 7.
- > Usage: `$trihard`
- > Aliases: `$TriHard`

### $tuck

- Tuck someone to bed.
- > Usage: `$tuck <user>`

## Fun

### $bibleverse

- Looks up a bible verse.
- > Usage: `$bibleverse <book> <chapter:verse>`
- > Aliases: `$bv`

### $cock

- Tells you the length :)
- > Usage: `$cock [user]`
- > Aliases: `$kok`

### $fortune

- Replies with a fortune. Fortunes from https://github.com/bmc/fortunes
- > Usage: `$fortune`

### $iq

- Tells you someone's IQ
- > Usage: `$iq [user]`

### $ship

- Tells you the compatibility of two people.
- > Usage: `$ship <first-person> <second-person>`

## Gamba

### $accept

- Accepts a duel.
- > Usage: `$accept`

### $decline

- Declines a duel.
- > Usage: `$decline`

### $duel

- Duels another chatter. They have 30 seconds to accept or decline.
- > Usage: `$duel <user> <amount>`
- > Per-user cooldown: `5s`

### $givepoints

- Give points to another chatter.
- > Usage: `$givepoints <user> <amount>`
- > Aliases: `$gp`

### $points

- Checks how many points someone has.
- > Usage: `$points [user]`
- > Aliases: `$p`

### $roulette

- Roulettes some points.
- > Usage: `$roulette <amount|percent%|all>`
- > Per-user cooldown: `5s`
- > Aliases: `$r`

## Kick

### $kickislive

- Replies with whether the Kick channel is currently live.
- > Usage: `$kickislive <channel>`
- > Aliases: `$kislive`

### $kicktitle

- Replies with the title of the Kick channel. Currently only works if the channel is live.
- > Usage: `$kicktitle <channel>`
- > Aliases: `$ktitle`

## Moderation

### $vanish

- Times you out for 1 second.
- > Usage: `$vanish`

## Twitch

### $banreason

- Replies with the reason someone was banned on Twitch.
- > Usage: `$banreason <user>`
- > Aliases: `$br`

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

### $subage

- Checks the length that someone has been subscribed to a channel on Twitch.
- > Usage: `$subage <user> <channel>`
- > Aliases: `$sa`, `$sublength`

### $title

- Replies with a channel's title. If no channel is provided, the current channel will be used.
- > Usage: `$title [channel]`

### $verifiedbot

- Replies whether a user is a verified bot. Currently offline due to changes on Twitch's end.
- > Usage: `$verifiedbot [user]`
- > Aliases: `$vb`

### $verifiedbotquiet

- Replies whether a user is a verified bot, but responds quietly. Currently offline due to changes on Twitch's end.
- > Usage: `$verifiedbotquiet [user]`
- > Aliases: `$verifiedbotq`, `$vbquiet`, `$vbq`

### $vips

- Replies with a channel's VIPs. If no channel is provided, the current channel will be used.
- > Usage: `$vips [channel]`
