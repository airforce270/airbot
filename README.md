# airbot

High-performance multi-platform utility bot written in Go.

Currently under development.

Support planned for many features (gamba, etc.) and platforms (Twitch, Discord, etc.).

## Running

To run the bot:

1. [Install Docker](https://docs.docker.com/get-docker/)
1. Clone the repository and change into it
1. Copy `config/config_example.json` to `config.json` in the main directory
1. Fill in the empty fields in `config.json`, notably API keys and usernames
1. Copy `.example.env` to `.env` in the main directory
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. Run `./start.sh`

### Production

To run in production (on a debian machine):

1. Clone the repository and change into it
1. Run `cd scripts`
1. Run `./setup-vm-debian.sh` to set up the environment
1. Fill in the empty fields in `config.json`, notably API keys and usernames
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. Reboot the machine
1. Run `./start-in-tmux.sh` to start the bot in a detached tmux session called `airbot`

If you want to view the logs or kill the bot, run `tmux attach -t airbot`
