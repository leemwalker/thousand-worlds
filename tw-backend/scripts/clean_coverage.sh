#!/usr/bin/env bash
set -euo pipefail

# Clean coverage artifacts
echo "Cleaning coverage artifacts..."

# Remove coverage files in root
rm -f coverage_unit.out coverage_integration.out coverage_combined.out coverage.out

# Remove coverage reports directory
rm -rf coverage-reports

# Clean Go test cache
echo "Cleaning Go test cache..."
go clean -testcache

echo "Cleanup complete."
