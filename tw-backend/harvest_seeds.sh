#!/bin/bash
set -euo pipefail

# Configuration
TOTAL_SEEDS=${1:-1000000000} # Default to 1 Billion seeds if not specified
PROFILE=${2:-modern}         # Default to 'modern' (71% ocean), use 'hadean' for 92% ocean
SEARCH_BATCH_SIZE=10000      # Number of seeds to search per batch
WORKERS=16                   # Increased workers for server environment (adjust if needed)
YEARS=10000000               # 10M years
OUTPUT_FILE="golden_seeds.json"
TEMP_LOG="search_progress.log"

echo "🌍 Starting Long-Running Golden Seed Harvest..."
echo "Target: Scan $TOTAL_SEEDS seeds for top 1% Earth-like matches"
echo "Profile: $PROFILE"
echo "Workers: $WORKERS"
echo "Output: $OUTPUT_FILE"
echo "Logging to: $TEMP_LOG"
echo "==================================================="

# Clear previous logs
rm -f "$TEMP_LOG"
echo "[]" > "$OUTPUT_FILE"

# Build the search tool first
echo "🔨 Building seed-search tool..."
go build -o bin/seed-search ./cmd/seed-search

start_time=$(date +%s)
current_seed=1

for (( i=0; i<$TOTAL_SEEDS; i+=$SEARCH_BATCH_SIZE )); do
    batch_end=$((current_seed + SEARCH_BATCH_SIZE - 1))
    echo "Processing Batch: Seeds $current_seed to $batch_end..."
    
    # Run user requested batch in JSON mode
    ./bin/seed-search \
        -start "$current_seed" \
        -count "$SEARCH_BATCH_SIZE" \
        -workers "$WORKERS" \
        -years "$YEARS" \
        -resolution 128 \
        -top 10 \
        -profile "$PROFILE" \
        -json > "batch_${i}.json"

    # Merge results into main file using jq (if installed) or simple appending
    # Here we assume a simple append for safety if jq isn't present, 
    # but strictly we'd want to merge the JSON arrays.
    # For robust JSON handling, we'll actually use a small Go helper or 
    # just rely on the user checking the individual batch files if they want raw data.
    # But for the "Top 1%", we'll filter right now.
    
    # Extract entries with Score > 80 (approx top 1% based on previous runs)
    # We use grep/awk to avoid jq dependency if possible, but JSON parsing is safer.
    # Let's trust the tool's JSON output.
    
    echo "  > Batch complete. Merging top candidates..."
    
    # Check if we found any golden seeds (>80 score)
    grep -o '{"seed":[^}]*"score":[8-9][0-9]\.[0-9]*[^}]*}' "batch_${i}.json" >> "all_golden_candidates.jsonl" || true
    
    # Clean up batch file
    rm "batch_${i}.json"
    
    current_seed=$((current_seed + SEARCH_BATCH_SIZE))
done

end_time=$(date +%s)
duration=$((end_time - start_time))

echo "==================================================="
echo "✅ Harvest Complete in ${duration}s"
echo "Candidates saved to: all_golden_candidates.jsonl"
echo "Sorting top 100 seeds..."

# Sort candidates by score (requires jq for reliable JSON sorting)
if command -v jq &> /dev/null; then
    jq -s 'sort_by(-.score) | .[0:100]' all_golden_candidates.jsonl > top_100_golden_seeds.json
    echo "🏆 Top 100 Seeds saved to: top_100_golden_seeds.json"
    
    # Display top 5
    echo "Top 5 Seeds:"
    jq -r '.[0:5] | .[] | "Seed: \(.seed) | Score: \(.score) | Ocean: \(.ocean_coverage)%"' top_100_golden_seeds.json
else
    echo "⚠️ jq not found. Manual sorting required on: all_golden_candidates.jsonl"
    echo "Sample candidate:"
    head -n 1 all_golden_candidates.jsonl
fi
