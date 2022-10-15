# airbot

High-performance multi-platform utility bot written in Go.

Currently under development.

Support planned for many features (gamba, etc.) and platforms (Twitch, Discord, etc.).

## Commands

All commands assume the `$` prefix, but note that the prefix is configurable
per-channel (in `config.json`).
To find out what the prefix is in a channel, use [$prefix](#$prefix).

Some commands include parameters.

If the parameter is wrapped in `<angle brackets>`, it's a **required** parameter.

If the it's wrapped in `[square brackets]`, it's an **optional** parameter.

### Echo

#### $commands

- Replies with a link to this page.
- > Usage: `$commands`
- > Alternate command: `$help`

#### $prefix

- Replies with the prefix in this channel.
- > Usage: `$prefix`

#### $TriHard

- Replies with ![TriHard](https://static-cdn.jtvnw.net/emoticons/v1/120232/1.0) 7.
- > Usage: `$TriHard`

### Twitch

#### $banreason

- Replies with the reason someone was banned on Twitch.
- > Usage: `$banreason [user]` (default: you)
- > Alternate command: `$br`

#### $currentgame

- Replies with the game that's currently being streamed on a channel.
- > Usage: `$currentgame [user]` (default: you)

#### $founders

- Replies with a channel's founders.
- > Usage: `$founders [user]` (default: you)

#### $mods

- Replies with a channel's mods.
- > Usage: `$mods [user]` (default: you)

#### $namecolor

- Replies with a user's name color.
- > Usage: `$namecolor [user]` (default: you)

#### $title

- Replies with a channel's title.
- > Usage: `$title [user]` (default: you)

#### $verifiedbot

- Replies whether a user is a verified bot.
- > Usage: `$verifiedbot [user]` (default: you)
- > Alternate command: `$vb`

#### $vips

- Replies with a channel's VIPs.
- > Usage: `$vips [user]` (default: you)

## Live instance

The bot is currently running on the Twitch account `af2bot`.

It's hosted on a Google Cloud Platform e2-micro Compute Engine instance.

### Add/remove bot

To add or remove the bot from a channel, edit the `config.json` file
and restart the bot.

## Development

### Running

To run the bot:

1. [Install Docker](https://docs.docker.com/get-docker/)
1. Clone the repository and change into it
1. Copy `config/config_example.json` to `config.json` in the main directory
1. Fill in the empty fields in `config.json`, notably API keys and usernames
1. Copy `.example.env` to `.env` in the main directory
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. Run `./start.sh`

### Running in production

To run in production (on a debian machine):

1. Clone the repository and change into it
1. Run `cd scripts`
1. Run `./setup-vm-debian.sh` to set up the environment
1. Fill in the empty fields in `config.json`, notably API keys and usernames
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. Reboot the machine
1. Run `./start-in-tmux.sh` to start the bot in a detached tmux session called `airbot`

If you want to view the logs or kill the bot, run `tmux attach -t airbot`
