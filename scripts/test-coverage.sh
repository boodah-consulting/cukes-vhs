#!/bin/bash

# Ensure the script fails if any command fails
set -e

# Create a directory for coverage reports
COVERAGE_DIR="./coverage"
mkdir -p "$COVERAGE_DIR"

# Set GOCOVERDIR for Go 1.24+ coverage
export GOCOVERDIR="$COVERAGE_DIR"

# Run Ginkgo tests with coverage
ginkgo -v --race --covermode=atomic --coverprofile="$COVERAGE_DIR/coverage.out" --skip-package=testdata ./...

# Generate HTML coverage report
go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"

echo "Coverage report generated in $COVERAGE_DIR"

