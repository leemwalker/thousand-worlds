#!/usr/bin/env bash
set -euo pipefail

# Generate coverage badge JSON for shields.io
# Usage: ./coverage_badge.sh [coverage_file]

COVERAGE_FILE="${1:-coverage_unit.out}"
WORK_DIR="tw-backend"

if [[ ! -f "$WORK_DIR/$COVERAGE_FILE" ]]; then
    echo '{"schemaVersion": 1, "label": "coverage", "message": "unknown", "color": "inactive"}'
    exit 0
fi

cd "$WORK_DIR" || exit 1
COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | tail -n 1 | awk '{print $3}' | sed 's/%//')
cd - > /dev/null || exit 1

if [[ -z "$COVERAGE" ]]; then
    echo '{"schemaVersion": 1, "label": "coverage", "message": "unknown", "color": "inactive"}'
    exit 0
fi

# Determine color
COLOR="red"
if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
    COLOR="green"
elif (( $(echo "$COVERAGE >= 60" | bc -l) )); then
    COLOR="yellow"
fi

echo "{\"schemaVersion\": 1, \"label\": \"coverage\", \"message\": \"${COVERAGE}%\", \"color\": \"${COLOR}\"}"
