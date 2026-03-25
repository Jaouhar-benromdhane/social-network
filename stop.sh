#!/usr/bin/env bash
# ─── Social Network — Stop (Linux/macOS) ─────────────────────────────────────

set -e

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

echo ""
echo " [Social Network] Stopping containers..."
echo ""

# Detect compose command
if docker compose version > /dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v docker-compose > /dev/null 2>&1; then
    COMPOSE="docker-compose"
else
    echo " [ERROR] docker compose not found."
    exit 1
fi

$COMPOSE -f docker-compose.yml stop

echo ""
echo " [OK] Containers stopped."
echo ""
