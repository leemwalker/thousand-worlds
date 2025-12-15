#!/bin/bash
# Quick update and rebuild script for development

set -e

echo "=== Thousand Worlds Update & Build ==="

# Pull latest code
echo "Pulling latest code..."
cd /home/walker/git/thousand-worlds
git pull

# Rebuild game server
echo "Rebuilding game-server..."
cd tw-backend
docker compose -f docker-compose.prod.yml build --no-cache game-server

# Restart game server
echo "Restarting game-server..."
docker compose -f docker-compose.prod.yml up -d game-server

# Show status
echo ""
echo "=== Build Complete ==="
docker compose -f docker-compose.prod.yml ps game-server

echo ""
echo "Verify at: http://10.0.0.17:3000"
echo "Or run: world reset && world simulate 100000"
