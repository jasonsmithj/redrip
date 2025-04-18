#!/bin/bash
set -e

# Setup script for tests to ensure environment is prepared
echo "Setting up test environment..."

# Create temporary config directory for tests that might need it
mkdir -p ~/.redrip
touch ~/.redrip/config.conf
cat > ~/.redrip/config.conf << EOF
redash_url = https://test-redash.com/api
api_key = test-api-key
sql_dir = /tmp/redrip-test
EOF

# Create a test SQL directory
mkdir -p /tmp/redrip-test

echo "Test environment setup complete!" 