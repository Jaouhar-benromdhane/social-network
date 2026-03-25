#!/usr/bin/env bash
# ─── Social Network — Start (Linux/macOS) ────────────────────────────────────

set -e

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

echo ""
echo " [Social Network] Starting containers..."
echo ""

# Detect compose command
if docker compose version > /dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v docker-compose > /dev/null 2>&1; then
    COMPOSE="docker-compose"
else
    echo " [ERROR] docker compose not found. Please install Docker Desktop."
    exit 1
fi

$COMPOSE -f docker-compose.yml up -d

echo ""
echo " [OK] Containers started!"
echo " Frontend : http://localhost:3000"
echo " Backend  : http://localhost:8080/api/health"
echo ""

docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep social-network || true
echo ""
