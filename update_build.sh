#!/bin/bash
# Quick update and rebuild script for development

set -euo pipefail

echo "=== Thousand Worlds Update & Build ==="

# Pull latest code
echo "Pulling latest code..."
cd /home/walker/git/thousand-worlds
git pull

# Install frontend dependencies (for new packages like Zod)
echo "Installing frontend dependencies..."
cd tw-frontend
npm install --legacy-peer-deps
cd ..

# Rebuild game server and frontend
echo "Rebuilding game-server and frontend..."
cd tw-backend
docker compose -f docker-compose.prod.yml build game-server frontend

# Build core-physics for K8s/Agones
echo "Building core-physics for Agones..."
docker build -t tw-backend/core-physics:latest -f Dockerfile.core-physics .

# Restart services
echo "Restarting services..."
docker compose -f docker-compose.prod.yml up -d game-server frontend

# Restart Agones Fleet
echo "Rolling out new physics engine..."
kubectl -n mud-world scale fleet world-simulation-fleet --replicas=0
kubectl -n mud-world scale fleet world-simulation-fleet --replicas=2

# Show status
echo ""
echo "=== Build Complete ==="
docker compose -f docker-compose.prod.yml ps game-server frontend

echo ""
echo "Verify at: http://10.0.0.17:3000"
echo "Or run: world reset && world simulate 100000"
