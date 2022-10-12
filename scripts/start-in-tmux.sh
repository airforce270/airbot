#!/usr/bin/env bash

tmux new-session -d -s airbot
tmux send-keys -t airbot './start-prod.sh' C-m
