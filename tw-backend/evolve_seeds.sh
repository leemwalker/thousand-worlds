#!/bin/bash
set -euo pipefail

# Configuration
INPUT_FILE="all_golden_candidates.jsonl"
OUTPUT_FILE="archean_survivors.jsonl"
TOP_N=${1:-20}
YEARS=${2:-1500000000} # 1.5 Billion Years (Archean)
PROFILE=${3:-archean}

if [ ! -f "$INPUT_FILE" ]; then
    echo "❌ Input file '$INPUT_FILE' not found. Run harvest_seeds.sh first."
    exit 1
fi

echo "🧬 Starting Evolutionary Selection: $PROFILE"
echo "----------------------------------------"
echo "Input: $INPUT_FILE"
echo "Target: Top $TOP_N candidates"
echo "Years: $YEARS"
echo "Profile: $PROFILE"
echo "----------------------------------------"

# Extract top N seeds
# Sort by score desc, take top N, extract just the seed ID
echo "🔍 Extracting top $TOP_N seeds..."
SEEDS=$(jq -s "sort_by(-.score) | .[0:$TOP_N] | .[].seed" "$INPUT_FILE")

echo "🧪 Testing candidates..."
echo "" > "$OUTPUT_FILE"

count=0
for seed in $SEEDS; do
    count=$((count+1))
    echo "  [$count/$TOP_N] Testing Seed $seed..."

    # Run simulation for next era
    # We use a batch size of 1 just to test this single seed
    ./bin/seed-search \
        -start "$seed" \
        -count 1 \
        -years "$YEARS" \
        -resolution 64 \
        -profile "$PROFILE" \
        -min-score 70.0 \
        -json > "temp_evo.json"

    # Save if valid
    if [ -s "temp_evo.json" ]; then
        if grep -q '"score":' "temp_evo.json"; then
             # Flatten and append
             if command -v jq &> /dev/null; then
                 jq -c '.[]' "temp_evo.json" >> "$OUTPUT_FILE"
             else
                 cat "temp_evo.json" >> "$OUTPUT_FILE"
             fi
             
             # Show result
             SCORE=$(jq -r '.[0].score' "temp_evo.json")
             echo "    ✅ Survivor! Score: $SCORE"
        else
             echo "    ❌ Failed benchmarks."
        fi
    fi
done

rm -f "temp_evo.json"

echo "----------------------------------------"
echo "🏁 Evolution Complete."
echo "Survivors saved to: $OUTPUT_FILE"

if [ -f "$OUTPUT_FILE" ] && command -v jq &> /dev/null; then
    echo ""
    echo "🏆 Top Survivors:"
    jq -s 'sort_by(-.score) | .[0:5] | .[] | "Seed: \(.seed) | Score: \(.score | floor)"' "$OUTPUT_FILE"
fi
