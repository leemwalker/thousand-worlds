#!/bin/bash

# Deployment script for Nobara Linux with GPU support

echo "Starting deployment for Nobara Linux..."

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed."
    exit 1
fi

# Check for Nvidia Container Toolkit
if ! docker info | grep -q "Runtimes.*nvidia"; then
    echo "Warning: Nvidia Container Toolkit not detected in Docker info."
    echo "Please ensure you have installed the nvidia-container-toolkit package."
    echo "On Nobara, this should be pre-installed or available via package manager."
    echo "Proceeding, but GPU features may fail..."
    sleep 3
fi

# Ensure .env exists
if [ ! -f .env ]; then
    echo "Creating .env from .env.template..."
    if [ -f .env.template ]; then
        cp .env.template .env
        echo "Please edit .env with your production secrets!"
        read -p "Press enter to continue after editing .env (or Ctrl+C to abort)"
    else
        echo "Error: .env.template not found."
        exit 1
    fi
fi

# Pull and Build
echo "Building and starting services..."
docker compose -f docker-compose.prod.yml up -d --build

echo "Deployment complete! Check status with 'docker compose -f docker-compose.prod.yml ps'"
