#!/bin/bash

# Function to kill background processes on exit
cleanup() {
    echo "Stopping services..."
    # Kill all child processes of this script
    pkill -P $$
    wait
    echo "Services stopped."
}

# Trap SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

echo "Starting Backend..."
(cd mud-platform-backend && ./scripts/dev.sh) &

echo "Starting Frontend..."
(cd mud-platform-client && npm run dev) &

# Wait for all background processes
wait
