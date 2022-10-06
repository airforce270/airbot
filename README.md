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
  - If running in production, run `./start-prod.sh` instead.
  This will use the latest image from the repo instead of building it.
