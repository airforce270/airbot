#!/usr/bin/env bash

echo 'Doing pre-install cleanup...'
git pull
sudo apt-get remove \
    docker
    docker-engine
    docker.io
    containerd
    runc

echo 'Updating packages...'
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get autoremove

echo 'Installing prerequisites...'
sudo apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

echo 'Adding apt repositories...'
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

echo 'Installing packages...'
sudo apt-get install \
    docker-ce
    docker-ce-cli
    containerd.io
    docker-compose-plugin

echo 'Creating config files...'
cp .example.env .env
cp config/config_example.json config.json

echo 'Setup complete.'
echo 'Fill in the necessary values in config.json and .env, then run ./start-prod.sh to start up the bot.'
