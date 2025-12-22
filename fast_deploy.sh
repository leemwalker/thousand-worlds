#!/bin/bash
set -euo pipefail

# Configuration
REMOTE_USER="walker"
REMOTE_HOST="10.0.0.17"
REMOTE_URI="ssh://${REMOTE_USER}@${REMOTE_HOST}"

echo "=== Fast Deploy to ${REMOTE_HOST} ==="
echo "Mode: Remote Docker Context (Building locally, running remotely)"

# Check for SSH access
if ! ssh -q -o BatchMode=yes -o ConnectTimeout=5 "${REMOTE_USER}@${REMOTE_HOST}" exit; then
    echo "Error: Cannot SSH to ${REMOTE_HOST} without password."
    echo "Please set up SSH keys: ssh-copy-id ${REMOTE_USER}@${REMOTE_HOST}"
    exit 1
fi

export DOCKER_HOST="${REMOTE_URI}"

echo "Deploying from tw-backend..."
cd tw-backend

# Build and Up in one go
# We only target game-server and frontend to avoid messing with database/infra
docker compose -f docker-compose.prod.yml up -d --build game-server frontend

echo "=== Deployment Complete ==="
echo "Logs:"
docker compose -f docker-compose.prod.yml logs --tail=20 game-server frontend
