#!/bin/bash

# Thousand Worlds - Pre-Deployment Validation Script
# Validates that all required environment variables are set before deployment

set -e

echo "========================================"
echo "  Pre-Deployment Security Validation"
echo "========================================"
echo ""

# Load .env if it exists
if [ -f .env ]; then
    echo "✓ Found .env file"
    set -a
    source .env
    set +a
else
    echo "⚠ No .env file found (using environment variables)"
fi

echo ""
echo "Checking required environment variables..."
echo ""

# Track validation status
VALIDATION_PASSED=true

# === REQUIRED SECRETS ===
check_var() {
    local var_name=$1
    local min_length=${2:-1}
    
    if [ -z "${!var_name}" ]; then
        echo "✗ $var_name: NOT SET"
        VALIDATION_PASSED=false
    elif [ ${#!var_name} -lt $min_length ]; then
        echo "✗ $var_name: TOO SHORT (minimum ${min_length} characters)"
        VALIDATION_PASSED=false
    else
        echo "✓ $var_name: SET (${#!var_name} characters)"
    fi
}

# Check critical secrets
check_var "JWT_SECRET" 32
check_var "POSTGRES_PASSWORD" 12
check_var "MONGO_PASSWORD" 12

# === CORS CONFIGURATION ===
echo ""
echo "Checking CORS configuration..."
if [ -z "$CORS_ALLOWED_ORIGINS" ]; then
    echo "⚠ CORS_ALLOWED_ORIGINS not set (will use default)"
elif [[ "$CORS_ALLOWED_ORIGINS" == *"*"* ]]; then
    echo "✗ CORS_ALLOWED_ORIGINS contains wildcard (*) - SECURITY RISK!"
    VALIDATION_PASSED=false
else
    echo "✓ CORS_ALLOWED_ORIGINS: $CORS_ALLOWED_ORIGINS"
fi

# === DATABASE CONFIGURATION ===
echo ""
echo "Checking database configuration..."
if [[ "$DATABASE_URL" == *"sslmode=disable"* ]]; then
    echo "⚠ DATABASE_URL has SSL disabled - consider enabling for production"
fi

# === SUMMARY ===
echo ""
echo "========================================"
if [ "$VALIDATION_PASSED" = true ]; then
    echo "✓ All validations passed!"
    echo "========================================"
    exit 0
else
    echo "✗ Validation failed!"
    echo "========================================"
    echo ""
    echo "Please fix the issues above before deploying."
    echo "See .env.template for reference."
    exit 1
fi
