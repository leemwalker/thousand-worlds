#!/bin/bash
set -euo pipefail

# Configuration
# Configuration
if [[ "${1:-}" == "forever" || "${1:-}" == "infinite" ]]; then
    INFINITE_MODE=true
    TOTAL_SEEDS=0 # Unused in infinite mode
    START_SEED=${2:-1}
    PROFILE=${3:-modern}
else
    INFINITE_MODE=false
    # Ensure first arg is a number, otherwise default to 1M
    if [[ "${1:-}" =~ ^[0-9]+$ ]]; then
        TOTAL_SEEDS=${1:-1000000}
    else
        echo "⚠️ Invalid seed count '$1', defaulting to 1,000,000"
        TOTAL_SEEDS=1000000
    fi
    START_SEED=${2:-1}
    PROFILE=${3:-modern}
fi

SEARCH_BATCH_SIZE=10000      # Number of seeds to search per batch
WORKERS=16                   # Increased workers for server environment (adjust if needed)
YEARS=0 # Zero-year simulation (Initial State Only)
RESOLUTION=32 # Lowest resolution for speed
STRATEGY="halftree" # Use optimized Half-Tree strategy
OUTPUT_FILE="golden_seeds.json"
TEMP_LOG="search_progress.log"

echo "🚀 Starting Golden Seed Harvest"
if [ "$INFINITE_MODE" = true ]; then
    echo "Mode: INFINITE RUN (Ctrl+C to stop)"
else
    echo "Target: $TOTAL_SEEDS seeds starting from $START_SEED"
fi
echo "Strategy: $STRATEGY | Profile: $PROFILE | Years: $YEARS | Resolution: $RESOLUTION"
echo "Workers: $WORKERS"
echo "Output: $OUTPUT_FILE"
echo "Logging to: $TEMP_LOG"
echo "==================================================="

mkdir -p seeds_found

# Clear previous logs
rm -f "$TEMP_LOG"
echo "[]" > "$OUTPUT_FILE"

# Build the tool
echo "🔨 Building seed-search tool..."
go build -o bin/seed-search ./cmd/seed-search
echo "✅ Build complete"

start_time=$(date +%s)

# Calculate number of batches if not in infinite mode
NUM_BATCHES=0
if [ "$INFINITE_MODE" = false ]; then
    NUM_BATCHES=$(( (TOTAL_SEEDS + SEARCH_BATCH_SIZE - 1) / SEARCH_BATCH_SIZE ))
    echo "📦 Processing in $NUM_BATCHES batches of $SEARCH_BATCH_SIZE seeds..."
else
    echo "📦 Processing in infinite batches of $SEARCH_BATCH_SIZE seeds..."
fi

BATCH_IDX=0
CURRENT_START=$START_SEED

while true; do
    if [ "$INFINITE_MODE" = false ] && [ $BATCH_IDX -ge $NUM_BATCHES ]; then
        break
    fi

    echo "  [Batch $((BATCH_IDX+1))] Scanning seeds $CURRENT_START to $((CURRENT_START + SEARCH_BATCH_SIZE - 1))..."
    
    # Run batch with min-score logging to stderr
    # Stdout goes to JSON file, Stderr goes to console (visible to user)
    ./bin/seed-search \
        -start "$CURRENT_START" \
        -count "$SEARCH_BATCH_SIZE" \
        -workers "$WORKERS" \
        -years "$YEARS" \
        -resolution "$RESOLUTION" \
        -strategy "$STRATEGY" \
        -profile "$PROFILE" \
        -min-score 80.0 \
        -json > "batch_${BATCH_IDX}.json"
    
    # Append high scoring seeds to master list
    if [ -s "batch_${BATCH_IDX}.json" ]; then
        # Check if we found any golden seeds (>80 score) to save permanently
        if grep -q '"score":' "batch_${BATCH_IDX}.json"; then
             if command -v jq &> /dev/null; then
                 # Flatten JSON array to JSON Lines (one object per line)
                 jq -c '.[]' "batch_${BATCH_IDX}.json" >> "all_golden_candidates.jsonl"
             else
                 # Fallback: Just append the raw array (messy but preserves data)
                 cat "batch_${BATCH_IDX}.json" >> "all_golden_candidates.jsonl"
             fi
        fi
        rm "batch_${BATCH_IDX}.json"
    fi

    # Prepare next batch
    CURRENT_START=$((CURRENT_START + SEARCH_BATCH_SIZE))
    BATCH_IDX=$((BATCH_IDX + 1))
done

end_time=$(date +%s)
duration=$((end_time - start_time))

echo "==================================================="
echo "✅ Harvest Complete in ${duration}s"
echo "Candidates saved to: all_golden_candidates.jsonl"

# Display top 5 from the collected file if it exists
if [ -f "all_golden_candidates.jsonl" ] && command -v jq &> /dev/null; then
    echo "Top Found Seeds:"
    # Process JSONL stream (one object per line)
    # Sort by score descending and take top 10
    jq -s 'sort_by(-.score) | .[0:10] | .[] | "Seed: \(.seed) | Score: \(.score | floor)"' all_golden_candidates.jsonl
fi
