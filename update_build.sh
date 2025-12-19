#!/bin/bash
# Quick update and rebuild script for development

set -e

echo "=== Thousand Worlds Update & Build ==="

# Pull latest code
echo "Pulling latest code..."
cd /home/walker/git/thousand-worlds
git pull

# Rebuild game server and frontend (clear build cache first)
echo "Clearing Docker build cache..."
docker builder prune -af
echo "Rebuilding game-server and frontend..."
cd tw-backend
docker compose -f docker-compose.prod.yml build --no-cache --pull game-server frontend

# Restart services
echo "Restarting services..."
docker compose -f docker-compose.prod.yml up -d game-server frontend

# Show status
echo ""
echo "=== Build Complete ==="
docker compose -f docker-compose.prod.yml ps game-server frontend

echo ""
echo "Verify at: http://10.0.0.17:3000"
echo "Or run: world reset && world simulate 100000"
