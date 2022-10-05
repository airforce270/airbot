#!/usr/bin/env bash

echo "[START] Building airbot..."
docker build -t airbot .
echo "[START] Built airbot."

echo "[START] Starting airbot..."
docker run -it --rm airbot
