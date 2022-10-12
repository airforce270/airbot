#!/usr/bin/env bash

echo '[Airbot Setup] Doing pre-install cleanup...'
git pull
sudo apt-get remove \
    docker \
    docker-engine \
    docker.io \
    containerd \
    runc

echo '[Airbot Setup] Updating packages...'
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get autoremove

echo '[Airbot Setup] Installing prerequisites...'
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

echo '[Airbot Setup] Adding apt repositories...'
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

echo '[Airbot Setup] Installing packages...'
sudo apt-get install -y \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-compose-plugin \
    tmux

echo '[Airbot Setup] Creating config files...'
cp .example.env .env
cp config/config_example.json config.json

echo '[Airbot Setup] Setup complete.'
echo '[Airbot Setup] Fill in the necessary values in config.json and .env, then run ./start-prod.sh to start up the bot.'
