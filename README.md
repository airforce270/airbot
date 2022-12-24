# airbot

<!-- markdownlint-disable-next-line line-length -->
[![CodeFactor](https://www.codefactor.io/repository/github/airforce270/airbot/badge)](https://www.codefactor.io/repository/github/airforce270/airbot) [![GoReportCard](https://goreportcard.com/badge/github.com/airforce270/airbot)](https://goreportcard.com/report/github.com/airforce270/airbot) [![Go version](https://img.shields.io/github/go-mod/go-version/airforce270/airbot.svg)](go.mod)
 [![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/airforce270/airbot)

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

To have the bot leave your channel, type `$leave` in your chat.

## Development

### Running locally

To run the bot locally:

1. [Install Docker](https://docs.docker.com/get-docker/)
1. Clone the repository and change into it
1. Copy `config/config_example.toml` to `config.toml` in the main directory
1. Fill in the empty fields in `config.toml`, notably API keys and usernames
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
1. Fill in the empty fields in `config.toml`, notably API keys and usernames
1. Set a value for `POSTGRES_PASSWORD` in `.env`
1. (optional): If running in a GCE container, follow
  [these instructions](https://docs.docker.com/config/containers/logging/configure/#configure-the-default-logging-driver)
  to set your default Docker `log-driver` to `gcplogs` (in
  `/etc/docker/daemon.json`) to send the Docker logs to Google Cloud Logging.
1. Reboot the machine
1. Run `./start-prod.sh` to start the bot

### Maintenance

To update the bot, run `./update.sh`.

To stop the bot, run `./stop-prod.sh`.

To connect to the bot's container while it's running, run `docker attach airbot-server-1`.

To connect to the database's container while it's running, run `docker attach airbot-database-1`.

To disconnect from a container, press `CTRL-p CTRL-q`.

#### Updating database

To update Postgres to a new major version (example is moving from version **14**
to **15**):

1. Make sure the version-specific migrator image is available [here](https://github.com/tianon/docker-postgres-upgrade)

1. Run:

    ```shell
    docker run --rm \
        --env-file=.env \
        -e POSTGRES_INITDB_ARGS="-U airbot" \
        -v airbot_postgres-14-data:/var/lib/postgresql/14/data \
        -v airbot_postgres-15-data:/var/lib/postgresql/15/data \
        tianon/postgres-upgrade:14-to-15
    ```

    - Note that the `POSTGRES_INITDB_ARGS` value should match the `PGUSER` value
      in your `.env` file.
    - If the upgrade fails, run `docker volume rm airbot_postgres-15-data` and
      try again.

1. Run:

    ```shell
    docker run --rm \
        -v airbot_postgres-14-data:/var/lib/postgresql/14/data \
        -v airbot_postgres-15-data:/var/lib/postgresql/15/data \
        busybox \
        cp /var/lib/postgresql/14/data/pg_hba.conf /var/lib/postgresql/15/data/pg_hba.conf
    ```

1. Update `docker-compose.yml` to use the new Postgres image
    - Under `services.database`, change `image` from `postgres:14` to `postgres:15`
1. Update `docker-compose.yml` to use the new volume
    - Under `volumes.postgres-data`, change `name` from `airbot_postgres-14-data`
      to `airbot_postgres-15-data`
