# airbot

[![CodeFactor](https://www.codefactor.io/repository/github/airforce270/airbot/badge)](https://www.codefactor.io/repository/github/airforce270/airbot)
[![GoReportCard](https://goreportcard.com/badge/github.com/airforce270/airbot)](https://goreportcard.com/report/github.com/airforce270/airbot)
[![Go version](https://img.shields.io/github/go-mod/go-version/airforce270/airbot.svg)](go.mod)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/airforce270/airbot)

High-performance multi-platform utility bot written in Go.

Currently under development.

Support planned for many features (gamba, etc.) and platforms (Twitch, Discord,
etc.).

## Commands

Commands are available on the [commands page](docs/commands.md).

## Live instance

The bot is currently running on the Twitch account `af2bot`.

It's hosted on an Oracle Cloud arm64 VM.

### Add/remove bot

To add the bot to your channel, type `$join` in `af2bot`'s chat.

To have the bot leave your channel, type `$leave` in your chat.

## Development

### Running locally

To run the bot locally:

1. [Install Go](https://go.dev/doc/install)
1. Clone the repository and `cd` into it
1. Copy `config/config_example.toml` to `config.toml` in the main directory
1. Fill in the empty fields in `config.toml`, notably API keys and usernames
1. Run `go run .`

### Tests

To run tests, run `go test ./...`

#### Documentation

Some documentation is generated.

After making changes, run `go generate ./...` from the main directory to
regenerate the docs.

## Running in production

To run in production:

1. [Install Go](https://go.dev/doc/install)
1. Clone the repository and `cd` into it
1. Run `git switch release`
1. Copy `config/config_example.toml` to `config.toml` in the main directory
1. Fill in the empty fields in `config.toml`, notably API keys and usernames
1. Reboot the machine
1. Run `go run .` to start the bot

Note: It's recommended to run the bot in a tmux or screen session so the bot
continues running when you disconnect from the machine.

By default, the SQLite database will be stored in the current directory. To
change where the data is stored, set `AIRBOT_SQLITE_DATA_DIR`, i.e.
`AIRBOT_SQLITE_DATA_DIR=/some/dir go run .`

### Maintenance

To update the bot, run `git pull`, then restart the bot.
