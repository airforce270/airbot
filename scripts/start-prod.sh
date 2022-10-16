#!/usr/bin/env bash

docker compose -f ../docker-compose.yml -f ../docker-compose.prod.yml pull
docker compose -f ../docker-compose.yml -f ../docker-compose.prod.yml up -it -d --no-build
