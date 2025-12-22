#!/bin/bash
set -euo pipefail

# Configuration
REMOTE_USER="walker"
REMOTE_HOST="10.0.0.17"
REMOTE_DIR="/home/walker/git/thousand-worlds"

echo "=== Fast Deploy to ${REMOTE_HOST} ==="
echo "Mode: Rsync + Remote Docker Compose"

# Check for SSH access
if ! ssh -q -o BatchMode=yes -o ConnectTimeout=5 "${REMOTE_USER}@${REMOTE_HOST}" exit; then
    echo "Error: Cannot SSH to ${REMOTE_HOST} without password."
    echo "Please set up SSH keys: ssh-copy-id ${REMOTE_USER}@${REMOTE_HOST}"
    exit 1
fi

echo "Syncing files to remote..."
# Sync current directory to remote, excluding heavy/ignored files
# -a: archive mode (preserves permissions, etc)
# -v: verbose
# -z: compress
# --delete: delete extraneous files on destination
rsync -avz --delete \
    --exclude '.git/' \
    --exclude 'node_modules/' \
    --exclude '.env' \
    --exclude '.DS_Store' \
    --exclude 'tmp/' \
    --exclude 'dist/' \
    --exclude 'coverage/' \
    ./ "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

echo "Triggering remote build & deploy..."
# We run 'npm install' on the remote host to update package-lock.json
# because the local machine might not have npm or have a stale lockfile.
ssh "${REMOTE_USER}@${REMOTE_HOST}" "cd ${REMOTE_DIR}/tw-frontend && \
    npm install --legacy-peer-deps && \
    cd ../tw-backend && \
    docker compose -f docker-compose.prod.yml up -d --build game-server frontend && \
    docker compose -f docker-compose.prod.yml logs --tail=20 game-server frontend"

echo "=== Deployment Complete ==="
