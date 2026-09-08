#!/usr/bin/env bash
set -e

PI_HOST="${PI_HOST:-matheus@100.92.173.88}"
PI_DIR="${PI_DIR:-/home/matheus/tesouro-backend-go}"

echo "=========================================="
echo "🚀 Deploying tesouro-backend-go to Raspberry Pi"
echo "Host: $PI_HOST"
echo "=========================================="

echo "📦 1. Pulling latest image from GHCR on Raspberry Pi..."
ssh "$PI_HOST" "cd $PI_DIR && docker compose pull liturgical-backend"

echo "🔄 2. Recreating container with new image..."
ssh "$PI_HOST" "cd $PI_DIR && docker compose up -d liturgical-backend"

echo "🧹 3. Pruning dangling Docker images to preserve SD/SSD space..."
ssh "$PI_HOST" "docker image prune -f"

echo "⏳ 4. Checking container status..."
sleep 2
ssh "$PI_HOST" "docker ps --filter name=liturgical-backend --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'"

echo "🩺 5. Smoke testing public production API..."
STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" "https://api.salvemaria.xyz/" || true)
if [ "$STATUS_CODE" = "200" ]; then
    echo "✅ API is healthy! HTTP status: $STATUS_CODE"
else
    echo "⚠️ Warning: Expected HTTP 200, got: $STATUS_CODE"
fi

echo "✨ Deploy finished successfully!"
