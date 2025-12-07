#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# THOUSAND WORLDS - TEST COVERAGE VERIFICATION SCRIPT
# ==============================================================================
# Automatically discovers, runs, and reports on both unit and integration test
# coverage across the entire codebase.
#
# Requirements:
#   - Unit tests: minimum 80% coverage
#   - Integration tests: 100% coverage
#
# Usage: ./verify_coverage.sh [options]
# Run with --help for full documentation
# ==============================================================================

# Script version
VERSION="1.0.0"

# Color codes (disabled if --no-color or non-terminal)
RED=""
GREEN=""
YELLOW=""
BLUE=""
CYAN=""
MAGENTA=""
BOLD=""
RESET=""

# Default configuration
WORK_DIR="mud-platform-backend"
UNIT_THRESHOLD=80
INTEGRATION_THRESHOLD=100
OUTPUT_DIR="coverage-reports"
UNIT_ONLY=false
INTEGRATION_ONLY=false
GENERATE_HTML=false
GENERATE_JSON=false
VERBOSE=false
NO_COLOR=false
FAIL_FAST=false
DRY_RUN=false
CUSTOM_THRESHOLD=""
TARGET_PACKAGES=""
LOG_FILE=""

# Coverage output files
UNIT_COVERAGE_FILE="coverage_unit.out"
INTEGRATION_COVERAGE_FILE="coverage_integration.out"
COMBINED_COVERAGE_FILE="coverage_combined.out"

# Track test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
TEST_DURATION=0

# ==============================================================================
# HELPER FUNCTIONS
# ==============================================================================

# Initialize colors
init_colors() {
    if [[ -t 1 ]] && [[ "$NO_COLOR" != "true" ]]; then
        RED='\033[0;31m'
        GREEN='\033[0;32m'
        YELLOW='\033[1;33m'
        BLUE='\033[0;34m'
        CYAN='\033[0;36m'
        MAGENTA='\033[0;35m'
        BOLD='\033[1m'
        RESET='\033[0m'
    fi
}

# Log message (also write to file if LOG_FILE is set)
log() {
    local message="$1"
    echo -e "$message"
    if [[ -n "$LOG_FILE" ]]; then
        echo -e "$message" | sed 's/\x1b\[[0-9;]*m//g' >> "$LOG_FILE"
    fi
}

# Print error message
error() {
    log "${RED}ERROR: $1${RESET}" >&2
}

# Print warning message
warn() {
    log "${YELLOW}WARNING: $1${RESET}" >&2
}

# Print info message
info() {
    log "${CYAN}INFO: $1${RESET}"
}

# Print success message
success() {
    log "${GREEN}✓ $1${RESET}"
}

# Print section header
section() {
    log ""
    log "${BOLD}${BLUE}═══════════════════════════════════════════════════════════════${RESET}"
    log "${BOLD}${BLUE}$1${RESET}"
    log "${BOLD}${BLUE}═══════════════════════════════════════════════════════════════${RESET}"
    log ""
}

# Print usage information
usage() {
    cat << EOF
${BOLD}Thousand Worlds - Test Coverage Verification Script${RESET}
Version: $VERSION

${BOLD}USAGE:${RESET}
    ./verify_coverage.sh [OPTIONS]

${BOLD}OPTIONS:${RESET}
    --unit-only              Run only unit tests
    --integration-only       Run only integration tests
    --html                   Generate HTML coverage report
    --json                   Output results in JSON format
    --verbose                Show detailed test output
    --packages <path>        Test specific package(s) only (e.g., ./internal/auth)
    --fail-under <n>         Set custom coverage threshold (overrides both unit and integration)
    --no-color               Disable colored output
    --fail-fast              Stop on first test failure
    --dry-run                Show what would be executed without running tests
    --log <file>             Write output to log file
    --help                   Show this help message

${BOLD}EXIT CODES:${RESET}
    0    All tests pass and coverage meets thresholds
    1    Tests failed
    2    Coverage below threshold
    3    Script error (invalid arguments, missing dependencies)

${BOLD}EXAMPLES:${RESET}
    # Run all tests with coverage verification
    ./verify_coverage.sh

    # Run only unit tests and generate HTML report
    ./verify_coverage.sh --unit-only --html

    # Test specific package
    ./verify_coverage.sh --packages ./mud-platform-backend/internal/auth

    # Generate JSON output for CI/CD
    ./verify_coverage.sh --json > coverage-report.json

    # Custom threshold for specific check
    ./verify_coverage.sh --unit-only --fail-under 75

${BOLD}CONFIGURATION:${RESET}
    Optional .coveragerc file can be placed at repository root.
    See README_COVERAGE.md for configuration format.

EOF
}

# Check prerequisites
check_prerequisites() {
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        log "Please install Go 1.21 or higher: https://golang.org/dl/"
        exit 3
    fi

    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    info "Found Go version: $go_version"

    if [[ ! -d "$WORK_DIR" ]]; then
        error "Backend directory not found: $WORK_DIR"
        log "This script should be run from the repository root."
        exit 3
    fi
}

# Load optional configuration file
load_config() {
    local config_file=".coveragerc"
    if [[ -f "$config_file" ]]; then
        info "Loading configuration from $config_file"
        # Simple YAML parsing for our needs (could be improved with yq)
        if grep -q "unit_tests:" "$config_file" 2>/dev/null; then
            local unit_val
            unit_val=$(grep "unit_tests:" "$config_file" | awk '{print $2}')
            if [[ -n "$unit_val" ]]; then
                UNIT_THRESHOLD=$unit_val
            fi
        fi
        if grep -q "integration_tests:" "$config_file" 2>/dev/null; then
            local int_val
            int_val=$(grep "integration_tests:" "$config_file" | awk '{print $2}')
            if [[ -n "$int_val" ]]; then
                INTEGRATION_THRESHOLD=$int_val
            fi
        fi
    fi
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit-only)
                UNIT_ONLY=true
                shift
                ;;
            --integration-only)
                INTEGRATION_ONLY=true
                shift
                ;;
            --html)
                GENERATE_HTML=true
                shift
                ;;
            --json)
                GENERATE_JSON=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --packages)
                TARGET_PACKAGES="$2"
                shift 2
                ;;
            --fail-under)
                CUSTOM_THRESHOLD="$2"
                shift 2
                ;;
            --no-color)
                NO_COLOR=true
                shift
                ;;
            --fail-fast)
                FAIL_FAST=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --log)
                LOG_FILE="$2"
                shift 2
                ;;
            --help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 3
                ;;
        esac
    done

    # Apply custom threshold if set
    if [[ -n "$CUSTOM_THRESHOLD" ]]; then
        UNIT_THRESHOLD=$CUSTOM_THRESHOLD
        INTEGRATION_THRESHOLD=$CUSTOM_THRESHOLD
    fi

    # Validate options
    if [[ "$UNIT_ONLY" == "true" ]] && [[ "$INTEGRATION_ONLY" == "true" ]]; then
        error "Cannot specify both --unit-only and --integration-only"
        exit 3
    fi
}

# Discover Go packages to test
discover_packages() {
    local base_dir="$1"
    local packages
    
    if [[ -n "$TARGET_PACKAGES" ]]; then
        # Use specified packages, stripping mud-platform-backend/ prefix if present
        packages=$(echo "$TARGET_PACKAGES" | sed 's|^mud-platform-backend/||' | sed 's|^\./mud-platform-backend/||')
    else
        # Discover all packages with tests
        cd "$base_dir" || exit 3
        packages=$(go list ./... 2>/dev/null | grep -v vendor || true)
        cd - > /dev/null || exit 3
    fi

    echo "$packages"
}

# Run unit tests with coverage
run_unit_tests() {
    section "RUNNING UNIT TESTS"
    
    local packages
    packages=$(discover_packages "$WORK_DIR")
    
    if [[ -z "$packages" ]]; then
        warn "No packages found to test"
        return 1
    fi

    local test_cmd="go test -short -coverprofile=$UNIT_COVERAGE_FILE -covermode=atomic"
    
    if [[ "$FAIL_FAST" == "true" ]]; then
        test_cmd="$test_cmd -failfast"
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_cmd="$test_cmd -v"
    fi

    test_cmd="$test_cmd $packages"

    if [[ "$DRY_RUN" == "true" ]]; then
        info "Would run: cd $WORK_DIR && $test_cmd"
        return 0
    fi

    info "Running unit tests..."
    if [[ "$VERBOSE" == "true" ]]; then
        info "Command: $test_cmd"
    fi

    local start_time
    start_time=$(date +%s)

    cd "$WORK_DIR" || exit 3
    
    local test_output
    local test_exit_code=0
    if [[ "$VERBOSE" == "true" ]]; then
        eval "$test_cmd" || test_exit_code=$?
    else
        test_output=$(eval "$test_cmd" 2>&1) || test_exit_code=$?
    fi

    local end_time
    end_time=$(date +%s)
    TEST_DURATION=$((end_time - start_time))

    cd - > /dev/null || exit 3

    # Parse test results
    if [[ $test_exit_code -eq 0 ]]; then
        success "Unit tests completed successfully"
    else
        error "Unit tests failed"
        if [[ "$VERBOSE" != "true" ]] && [[ -n "$test_output" ]]; then
            echo "$test_output"
        fi
        return 1
    fi

    return 0
}

# Run integration tests with coverage
run_integration_tests() {
    section "RUNNING INTEGRATION TESTS"
    
    local packages
    packages=$(discover_packages "$WORK_DIR")
    
    if [[ -z "$packages" ]]; then
        warn "No packages found to test"
        return 1
    fi

    # Integration tests are identified by *_integration_test.go files
    # and they skip themselves when testing.Short() is true
    local test_cmd="go test -run Integration -coverprofile=$INTEGRATION_COVERAGE_FILE -covermode=atomic"
    
    if [[ "$FAIL_FAST" == "true" ]]; then
        test_cmd="$test_cmd -failfast"
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_cmd="$test_cmd -v"
    fi

    test_cmd="$test_cmd $packages"

    if [[ "$DRY_RUN" == "true" ]]; then
        info "Would run: cd $WORK_DIR && $test_cmd"
        return 0
    fi

    info "Running integration tests..."
    if [[ "$VERBOSE" == "true" ]]; then
        info "Command: $test_cmd"
    fi

    local start_time
    start_time=$(date +%s)

    cd "$WORK_DIR" || exit 3
    
    local test_output
    local test_exit_code=0
    if [[ "$VERBOSE" == "true" ]]; then
        eval "$test_cmd" || test_exit_code=$?
    else
        test_output=$(eval "$test_cmd" 2>&1) || test_exit_code=$?
    fi

    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    TEST_DURATION=$((TEST_DURATION + duration))

    cd - > /dev/null || exit 3

    # Parse test results
    if [[ $test_exit_code -eq 0 ]]; then
        success "Integration tests completed successfully"
    else
        error "Integration tests failed"
        if [[ "$VERBOSE" != "true" ]] && [[ -n "$test_output" ]]; then
            echo "$test_output"
        fi
        return 1
    fi

    return 0
}

# Calculate coverage from coverage file
calculate_coverage() {
    local coverage_file="$1"
    local coverage_output
    
    if [[ ! -f "$WORK_DIR/$coverage_file" ]]; then
        echo "0.0"
        return
    fi

    cd "$WORK_DIR" || exit 3
    coverage_output=$(go tool cover -func="$coverage_file" 2>/dev/null | tail -n 1 | awk '{print $3}' | sed 's/%//')
    cd - > /dev/null || exit 3

    if [[ -z "$coverage_output" ]]; then
        echo "0.0"
    else
        echo "$coverage_output"
    fi
}

# Get per-package coverage
get_package_coverage() {
    local coverage_file="$1"
    
    if [[ ! -f "$WORK_DIR/$coverage_file" ]]; then
        return
    fi

    cd "$WORK_DIR" || exit 3
    go tool cover -func="$coverage_file" 2>/dev/null | grep -E '^.*\.go:' | \
        awk '{pkg=$1; sub(/\/[^\/]+\.go:.*/, "", pkg); cov[pkg]+=$NF; count[pkg]++} 
             END {for (p in cov) printf "%s %.1f\n", p, cov[p]/count[p]}' | \
        sort
    cd - > /dev/null || exit 3
}

# Display coverage results
display_coverage_results() {
    local test_type="$1"
    local coverage_file="$2"
    local threshold="$3"
    
    section "$test_type TEST COVERAGE RESULTS"
    
    if [[ ! -f "$WORK_DIR/$coverage_file" ]]; then
        warn "No coverage data found for $test_type tests"
        return 1
    fi

    # Display header
    printf "${BOLD}%-60s %10s %10s${RESET}\n" "Package" "Coverage" "Status"
    printf "%s\n" "────────────────────────────────────────────────────────────────────────────"

    local overall_coverage
    overall_coverage=$(calculate_coverage "$coverage_file")
    
    # Display per-package coverage
    local package_data
    package_data=$(get_package_coverage "$coverage_file")
    
    local below_threshold=false
    local failed_packages=()
    
    while IFS= read -r line; do
        if [[ -z "$line" ]]; then
            continue
        fi
        
        local pkg
        local cov
        pkg=$(echo "$line" | awk '{print $1}')
        cov=$(echo "$line" | awk '{print $2}')
        
        # Determine status
        local status
        local color
        if (( $(echo "$cov >= $threshold" | bc -l) )); then
            status="✓ PASS"
            color="$GREEN"
        else
            status="✗ FAIL"
            color="$RED"
            below_threshold=true
            local needed
            needed=$(echo "$threshold - $cov" | bc -l)
            failed_packages+=("$pkg: ${cov}% (need ${needed}% more)")
        fi
        
        printf "${color}%-60s %9.1f%% %10s${RESET}\n" "$pkg" "$cov" "$status"
    done <<< "$package_data"

    printf "%s\n" "────────────────────────────────────────────────────────────────────────────"
    
    # Overall coverage
    local overall_status
    local overall_color
    if (( $(echo "$overall_coverage >= $threshold" | bc -l) )); then
        overall_status="✓ PASS"
        overall_color="$GREEN"
    else
        overall_status="✗ FAIL"
        overall_color="$RED"
        below_threshold=true
    fi
    
    printf "${BOLD}%-60s %9.1f%% %10s${RESET}\n" "Overall $test_type Coverage" "$overall_coverage" "$overall_status"
    echo ""

    # Store failed packages for summary
    if [[ "$below_threshold" == "true" ]]; then
        echo "${failed_packages[@]:-}" > "/tmp/failed_packages_${test_type}.txt"
        return 1
    fi
    
    return 0
}

# Generate HTML coverage report
generate_html_report() {
    if [[ "$GENERATE_HTML" != "true" ]]; then
        return
    fi

    section "GENERATING HTML COVERAGE REPORT"

    mkdir -p "$OUTPUT_DIR"

    cd "$WORK_DIR" || exit 3

    if [[ -f "$UNIT_COVERAGE_FILE" ]]; then
        info "Generating HTML report for unit tests..."
        go tool cover -html="$UNIT_COVERAGE_FILE" -o="../$OUTPUT_DIR/coverage_unit.html"
        success "Unit test HTML report: $OUTPUT_DIR/coverage_unit.html"
    fi

    if [[ -f "$INTEGRATION_COVERAGE_FILE" ]]; then
        info "Generating HTML report for integration tests..."
        go tool cover -html="$INTEGRATION_COVERAGE_FILE" -o="../$OUTPUT_DIR/coverage_integration.html"
        success "Integration test HTML report: $OUTPUT_DIR/coverage_integration.html"
    fi

    cd - > /dev/null || exit 3
}

# Generate JSON coverage report
generate_json_report() {
    if [[ "$GENERATE_JSON" != "true" ]]; then
        return
    fi

    local unit_coverage="0.0"
    local integration_coverage="0.0"

    if [[ -f "$WORK_DIR/$UNIT_COVERAGE_FILE" ]]; then
        unit_coverage=$(calculate_coverage "$UNIT_COVERAGE_FILE")
    fi

    if [[ -f "$WORK_DIR/$INTEGRATION_COVERAGE_FILE" ]]; then
        integration_coverage=$(calculate_coverage "$INTEGRATION_COVERAGE_FILE")
    fi

    cat > "$OUTPUT_DIR/coverage_report.json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "unit_tests": {
    "coverage": $unit_coverage,
    "threshold": $UNIT_THRESHOLD,
    "passed": $(if (( $(echo "$unit_coverage >= $UNIT_THRESHOLD" | bc -l) )); then echo "true"; else echo "false"; fi)
  },
  "integration_tests": {
    "coverage": $integration_coverage,
    "threshold": $INTEGRATION_THRESHOLD,
    "passed": $(if (( $(echo "$integration_coverage >= $INTEGRATION_THRESHOLD" | bc -l) )); then echo "true"; else echo "false"; fi)
  },
  "summary": {
    "total_duration_seconds": $TEST_DURATION,
    "all_passed": $(if (( $(echo "$unit_coverage >= $UNIT_THRESHOLD" | bc -l) )) && (( $(echo "$integration_coverage >= $INTEGRATION_THRESHOLD" | bc -l) )); then echo "true"; else echo "false"; fi)
  }
}
EOF

    success "JSON report generated: $OUTPUT_DIR/coverage_report.json"
    
    if [[ "$GENERATE_JSON" == "true" ]] && [[ "$GENERATE_HTML" != "true" ]]; then
        cat "$OUTPUT_DIR/coverage_report.json"
    fi
}

# Display final summary
display_summary() {
    local unit_passed="$1"
    local integration_passed="$2"
    local unit_coverage="$3"
    local integration_coverage="$4"

    section "SUMMARY"

    # Unit tests summary
    if [[ "$INTEGRATION_ONLY" != "true" ]]; then
        if [[ "$unit_passed" == "true" ]]; then
            success "Unit Tests:        PASS (${unit_coverage}% >= ${UNIT_THRESHOLD}%)"
        else
            log "${RED}✗ Unit Tests:        FAIL (${unit_coverage}% < ${UNIT_THRESHOLD}%)${RESET}"
        fi
    fi

    # Integration tests summary
    if [[ "$UNIT_ONLY" != "true" ]]; then
        if [[ "$integration_passed" == "true" ]]; then
            success "Integration Tests: PASS (${integration_coverage}% >= ${INTEGRATION_THRESHOLD}%)"
        else
            log "${RED}✗ Integration Tests: FAIL (${integration_coverage}% < ${INTEGRATION_THRESHOLD}%)${RESET}"
        fi
    fi

    log ""
    log "Total Test Duration: ${TEST_DURATION}s"
    log ""
    log "${BOLD}${BLUE}═══════════════════════════════════════════════════════════════${RESET}"

    # Overall result
    if [[ "$unit_passed" == "true" ]] && [[ "$integration_passed" == "true" ]]; then
        log "${BOLD}${GREEN}RESULT: ✓ ALL CHECKS PASSED${RESET}"
        log "${BOLD}${BLUE}═══════════════════════════════════════════════════════════════${RESET}"
        return 0
    else
        log "${BOLD}${RED}RESULT: ✗ COVERAGE THRESHOLD NOT MET${RESET}"
        log "${BOLD}${BLUE}═══════════════════════════════════════════════════════════════${RESET}"
        
        # Show failed packages
        if [[ -f "/tmp/failed_packages_Unit.txt" ]]; then
            log ""
            log "The following packages are below the ${UNIT_THRESHOLD}% unit test coverage threshold:"
            log ""
            while IFS= read -r line; do
                log "  ${RED}$line${RESET}"
            done < "/tmp/failed_packages_Unit.txt"
        fi
        
        if [[ -f "/tmp/failed_packages_Integration.txt" ]]; then
            log ""
            log "The following packages are below the ${INTEGRATION_THRESHOLD}% integration test coverage threshold:"
            log ""
            while IFS= read -r line; do
                log "  ${RED}$line${RESET}"
            done < "/tmp/failed_packages_Integration.txt"
        fi
        
        log ""
        log "Run with --verbose to see uncovered lines."
        log "Run with --html to generate detailed coverage report."
        log ""
        return 2
    fi
}

# Cleanup temporary files
cleanup() {
    rm -f "/tmp/failed_packages_Unit.txt" "/tmp/failed_packages_Integration.txt" 2>/dev/null || true
}

# ==============================================================================
# MAIN EXECUTION
# ==============================================================================

main() {
    # Parse arguments first (before colors, so --no-color works)
    parse_args "$@"
    
    # Initialize colors
    init_colors

    # Print header
    section "THOUSAND WORLDS - TEST COVERAGE VERIFICATION"
    
    # Check prerequisites
    check_prerequisites
    
    # Load configuration
    load_config

    # Show configuration
    if [[ "$DRY_RUN" == "true" ]]; then
        info "DRY RUN MODE - No tests will be executed"
    fi
    info "Unit test threshold: ${UNIT_THRESHOLD}%"
    info "Integration test threshold: ${INTEGRATION_THRESHOLD}%"

    local unit_passed=true
    local integration_passed=true
    local unit_coverage="0.0"
    local integration_coverage="0.0"
    local tests_failed=false

    # Run unit tests
    if [[ "$INTEGRATION_ONLY" != "true" ]]; then
        if ! run_unit_tests; then
            tests_failed=true
        fi
        
        if [[ "$DRY_RUN" != "true" ]]; then
            unit_coverage=$(calculate_coverage "$UNIT_COVERAGE_FILE")
            if ! display_coverage_results "Unit" "$UNIT_COVERAGE_FILE" "$UNIT_THRESHOLD"; then
                unit_passed=false
            fi
        fi
    fi

    # Run integration tests
    if [[ "$UNIT_ONLY" != "true" ]]; then
        if ! run_integration_tests; then
            tests_failed=true
        fi
        
        if [[ "$DRY_RUN" != "true" ]]; then
            integration_coverage=$(calculate_coverage "$INTEGRATION_COVERAGE_FILE")
            if ! display_coverage_results "Integration" "$INTEGRATION_COVERAGE_FILE" "$INTEGRATION_THRESHOLD"; then
                integration_passed=false
            fi
        fi
    fi

    # Generate reports
    if [[ "$DRY_RUN" != "true" ]]; then
        generate_html_report
        generate_json_report
    fi

    # Display summary
    if [[ "$DRY_RUN" != "true" ]]; then
        if ! display_summary "$unit_passed" "$integration_passed" "$unit_coverage" "$integration_coverage"; then
            cleanup
            exit 2
        fi
    else
        success "Dry run completed"
    fi

    # Check if tests failed
    if [[ "$tests_failed" == "true" ]]; then
        cleanup
        exit 1
    fi

    cleanup
    exit 0
}

# Run main function
main "$@"
