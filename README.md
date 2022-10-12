# airbot

High-performance multi-platform utility bot written in Go.

Currently under development.

Support planned for many features (gamba, etc.) and platforms (Twitch, Discord, etc.).

## Running

To run the bot:

- [Install Docker](https://docs.docker.com/get-docker/)
- Clone the repository and change into it
- Copy `config/config_example.json` to `config.json` in the main directory
- Fill in the empty fields in `config.json`, notably API keys and usernames
- Copy `.example.env` to `.env` in the main directory
- Set a value for `POSTGRES_PASSWORD` in `.env`
- Run `./start.sh`

### Production

To run in production (on a debian machine):

- Clone the repository and change into it
- Run `cd scripts`
- Run `./setup-vm-debian.sh` to set up the environment
- Run `./start-in-tmux.sh` to start the bot in a detached tmux session called `airbot`

If you want to view the logs or kill the bot, run `tmux attach -t airbot`
