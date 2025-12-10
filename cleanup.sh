#!/bin/bash

# Thousand Worlds - Cleanup Script
# Kills running processes and manages log files

echo "Killing any existing processes..."

# Kill game-server process if running
pkill -9 game-server 2>/dev/null || true

# Kill processes on ports (backend 8080, frontend 5173)
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
lsof -ti:5173 | xargs kill -9 2>/dev/null || true

# Backup and truncate server log if it exists
if [ -f "tw-backend/server.log" ]; then
    DATE=$(date +%Y%m%d%H%M%S)
    cp tw-backend/server.log tw-backend/server.log.backup.$DATE
    truncate -s 0 tw-backend/server.log
    echo "Server log backed up and truncated"
fi

# Clean up PID files
rm -f logs/backend.pid logs/frontend.pid 2>/dev/null || true

echo "Cleanup complete"
