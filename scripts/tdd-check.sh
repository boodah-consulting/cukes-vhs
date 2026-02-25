#!/bin/bash

# TDD Check Script
#
# Verifies that test files are staged along with production code
# to enforce Test-Driven Development practices

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Get staged production Go files (excluding tests and BDD test infrastructure)
# The features/ directory contains BDD step definitions which are test code, not production code
STAGED_PROD_FILES=$(git diff --cached --name-only --diff-filter=AM | grep '\.go$' | grep -v '_test\.go$' | grep -v '^features/' || true)

if [ -z "$STAGED_PROD_FILES" ]; then
    # No production Go files staged, TDD check not applicable
    exit 0
fi

# Check if corresponding test files are staged
TDD_VIOLATIONS=0
MISSING_TESTS=()

while IFS= read -r prod_file; do
    # Derive test file name
    test_file="${prod_file%.go}_test.go"
    
    # Check if test file is staged
    if ! git diff --cached --name-only | grep -q "^${test_file}$"; then
        # Check if test file exists at all
        if [ -f "$test_file" ]; then
            echo -e "${YELLOW}⚠️  Production file modified but test not updated: ${prod_file}${NC}"
            MISSING_TESTS+=("$prod_file")
            TDD_VIOLATIONS=$((TDD_VIOLATIONS+1))
        else
            echo -e "${RED}⚠️  Production file added without test: ${prod_file}${NC}"
            MISSING_TESTS+=("$prod_file")
            TDD_VIOLATIONS=$((TDD_VIOLATIONS+1))
        fi
    fi
done <<< "$STAGED_PROD_FILES"

# Report results
if [ $TDD_VIOLATIONS -eq 0 ]; then
    echo -e "${GREEN}✅ TDD check passed: Test files staged with production code${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}⚠️  TDD WARNING: Production code without test changes${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "The following production files are staged without corresponding test changes:"
    echo ""
    for file in "${MISSING_TESTS[@]}"; do
        echo "  - $file"
    done
    echo ""
    echo -e "${YELLOW}TDD Practice: Tests should be written BEFORE implementation${NC}"
    echo ""
    echo "Options:"
    echo "  1. Add/update test files and stage them"
    echo "  2. If this is truly test-only work, continue"
    echo "  3. Override with: SKIP_TDD_CHECK=1 git commit"
    echo ""
    echo "See: docs/rules/senior-engineer-guidelines.md (Red-Green-Refactor)"
    echo ""
    
    # Return warning code (non-zero but not failure)
    # The pre-commit hook can decide whether to block
    exit 10
fi
