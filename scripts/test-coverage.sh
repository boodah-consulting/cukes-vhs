#!/usr/bin/env bash
# Test coverage report for cukes-vhs
# Non-interactive — generates coverage data and summary

set -euo pipefail

echo "Running tests with coverage..."
ginkgo -v --race --skip-package=testdata --coverprofile=coverage.out --covermode=atomic ./...

echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "Coverage report generated: coverage.out"
echo "View HTML report: go tool cover -html=coverage.out"
