# airbot

High-performance multi-platform utility bot written in Go.

Currently under development.

Support planned for many features (gamba, etc.) and platforms (Twitch, Discord, etc.).

## Commands

Commands are available on the [commands page](docs/commands.md).

## Live instance

The bot is currently running on the Twitch account `af2bot`.

It's hosted on a Google Cloud Platform e2-micro Compute Engine instance.

### Add/remove bot

To add the bot to your channel, type `$join` in `af2bot`'s chat.

WARNING! The prefix will be `$` and there's currently no way to change that yourself.
(coming soon!)

To have the bot leave your channel, type `$leave` in your chat.

## Development

### Running locally

To run the bot locally:

1. [Install Docker](https://docs.docker.com/get-docker/)
1. Clone the repository and change into it
1. Copy `config/config_example.json` to `config.json` in the main directory
1. Fill in the empty fields in `config.json`, notably API keys and usernames
1. Copy `.example.env` to `.env` in the main directory
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. Run `./start.sh`

#### Documentation

Some documentation is generated.

After making changes, run `./gen-docs.sh` from the main directory to regenerate
the docs.

## Running in production

To run in production (on a debian machine):

1. Clone the repository and change into it
1. Run `cd scripts`
1. Run `./setup-vm-debian.sh` to set up the environment
1. Fill in the empty fields in `config.json`, notably API keys and usernames
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. Reboot the machine
1. Run `./start-prod.sh` to start the bot

### Maintenance

To update the bot, run `./update.sh`.

To stop the bot, run `./stop-prod.sh`.

To connect to the bot's container while it's running, run `docker attach airbot-server-1`.

To connect to the database's container while it's running, run `docker attach airbot-database-1`.

To disconnect from a container, press `CTRL-p CTRL-q`.
