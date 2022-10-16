#!/usr/bin/env bash

git pull
docker compose -f ../docker-compose.yml -f ../docker-compose.prod.yml pull
docker compose -f ../docker-compose.yml -f ../docker-compose.prod.yml down
docker compose -f ../docker-compose.yml -f ../docker-compose.prod.yml up -it -d --no-build
