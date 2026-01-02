#!/bin/sh
set -e

cd /app

echo "Pulling latest images..."
docker compose --profile prod pull bot

echo "Restarting services..."
docker compose --profile prod up -d bot

echo "Cleaning up old images..."
docker image prune -f

echo "Deploy complete"
