[//]: # ( !!! DO NOT EDIT MANUALLY !!!  This is a generated file, any changes will be overwritten! )

# Commands

All commands assume the `$` prefix, but note that the prefix is configurable
per-channel (in `config.json`).
To find out what the prefix is in a channel, ask `what's airbot's prefix?`
in a chat.

Some commands include parameters.

If the parameter is wrapped in `<angle brackets>`, it's a **required** parameter.

If the it's wrapped in `[square brackets]`, it's an **optional** parameter.

## Admin

### $join

- Tells the bot to join your chat.
- > Usage: `$join`

### $joinother

- Tells the bot to join a chat.
- > Usage: `$joinother <channel>`
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

## Echo

### $commands

- Replies with a link to the commands.
- > Usage: `$commands`

### $TriHard

- Replies with TriHard 7.
- > Usage: `$TriHard`

## Twitch

### $banreason

- Replies with the reason someone was banned on Twitch.

### $currentgame

- Replies with the game that's currently being streamed on a channel.

### $founders

- Replies with a channel's founders.

### $logs

- Replies with a link to a Twitch user's logs in a channel.
- > Usage: `$logs <channel> <user>`

### $mods

- Replies with a channel's mods.
- > Usage: `$mods [user]`

### $namecolor

- Replies with a user's name color.
- > Usage: `$namecolor [user]`

### $title

- Replies with a channel's title.
- > Usage: `$title [user]`

### $verifiedbot

- Replies whether a user is a verified bot.
- > Usage: `$verifiedbot [user]`
- > Alternate commands: `$vb`

### $vips

- Replies with a channel's VIPs.
- > Usage: `$vips [user]`
