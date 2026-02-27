#!/bin/bash

set -e

echo "================================================"
echo "🔍 RULES COMPLIANCE CHECK"
echo "================================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

VIOLATIONS=0
WARNINGS=0

# Helper function for checks
check_pass() {
    echo -e "${GREEN}✅ Pass${NC}"
}

check_fail() {
    echo -e "${RED}❌ Fail${NC} - $1"
    VIOLATIONS=$((VIOLATIONS+1))
}

check_warn() {
    echo -e "${YELLOW}⚠️  Warning${NC} - $1"
    WARNINGS=$((WARNINGS+1))
}

# ============================================
# 1. CODE QUALITY CHECKS
# ============================================
echo "📋 CODE QUALITY"
echo "------------------------------------------------"

# Formatting
echo -n "Formatting (gofmt): "
UNFORMATTED=$(gofmt -l . 2>&1 | grep -v '^vendor/' | grep '\.go$' || true)
if [ -z "$UNFORMATTED" ]; then
    check_pass
else
    check_fail "Run: go fmt ./..."
    echo "  Unformatted files:"
    echo "$UNFORMATTED" | sed 's/^/    /'
fi

# Build
echo -n "Build: "
if go build ./... > /dev/null 2>&1; then
    check_pass
else
    check_fail "Fix build errors"
fi

# Tests
echo -n "Tests: "
if go test ./... > /dev/null 2>&1; then
    check_pass
else
    check_fail "Fix failing tests"
fi

# Race Conditions (CI only - too slow for local checks)
# Run `make test-race` manually if needed

# Vet
echo -n "Go Vet: "
if go vet ./... > /dev/null 2>&1; then
    check_pass
else
    check_fail "Fix vet warnings"
fi

# Staticcheck
echo -n "Staticcheck: "
if command -v staticcheck &> /dev/null; then
    if staticcheck ./... > /dev/null 2>&1; then
        check_pass
    else
        check_fail "Fix staticcheck warnings (run: staticcheck ./...)"
    fi
else
    check_warn "staticcheck not installed (run: go install honnef.co/go/tools/cmd/staticcheck@latest)"
fi

echo ""

# ============================================
# 2. TEST COVERAGE
# ============================================
echo "📊 TEST COVERAGE"
echo "------------------------------------------------"
echo ""
echo -e "${BLUE}Coverage Requirements:${NC}"
echo "  - Per-package (modified): >= 95% (enforced by pre-commit hook)"
echo "  - Project average:        >= 80% (warning threshold)"
echo ""

if command -v ginkgo &> /dev/null; then
    # Using Ginkgo
    # Exclude mock packages, test utilities, and packages with no statements from coverage calculation
    COVERAGE_OUTPUT=$(go test -cover ./... 2>/dev/null || true)
    if [ -n "$COVERAGE_OUTPUT" ]; then
        # Filter out mocks, testutil (but keep e2e), and packages with 0.0% or "no test files"
        COVERAGE=$(echo "$COVERAGE_OUTPUT" | \
            grep -v '/mocks' | \
            grep -v 'testutil[^/]' | \
            grep -v 'test_all_views' | \
            grep -v '\[no test' | \
            grep -oP 'coverage: \K[0-9.]+' | \
            awk '$1 > 0 {sum+=$1; count++} END {if(count>0) printf "%.4f", sum/count; else print 0}')
        COVERAGE_INT=$(printf "%.0f" "$COVERAGE")

        if [ "$COVERAGE_INT" -ge 80 ]; then
            echo -e "Project Average: ${GREEN}${COVERAGE}% ✅${NC}"
        elif [ "$COVERAGE_INT" -ge 70 ]; then
            echo -e "Project Average: ${YELLOW}${COVERAGE}% ⚠️${NC} (Target: 80%)"
            check_warn "Project average coverage below 80%"
        else
            echo -e "Project Average: ${RED}${COVERAGE}% ❌${NC} (Target: 80%)"
            check_fail "Project average coverage significantly below 80%"
        fi
        
        # Check for modules below 95% threshold (matches pre-commit requirement)
        echo ""
        echo "Packages below 95% coverage (pre-commit will block these if modified):"
        LOW_COVERAGE_MODULES=$(echo "$COVERAGE_OUTPUT" | \
            grep -v '/mocks' | \
            grep -v 'testutil[^/]' | \
            grep -v 'test_all_views' | \
            grep -v '\[no test' | \
            grep 'coverage:' | \
            awk '{
                match($0, /coverage: ([0-9.]+)%/, arr);
                if (arr[1]+0 < 95 && arr[1]+0 > 0) {
                    # Extract package name (first field after "ok")
                    gsub(/^ok[[:space:]]+/, "");
                    split($0, parts, /[[:space:]]/);
                    printf "  %s: %s%%\n", parts[1], arr[1]
                }
            }')
        
        if [ -n "$LOW_COVERAGE_MODULES" ]; then
            echo -e "${YELLOW}$LOW_COVERAGE_MODULES${NC}"
            echo ""
            echo -e "${BLUE}Tip: Add tests before modifying these packages.${NC}"
        else
            echo -e "  ${GREEN}All packages meet 95% threshold ✅${NC}"
        fi
    else
        check_warn "Could not calculate coverage"
    fi
else
    check_warn "Ginkgo not installed, skipping coverage check"
fi

echo ""

# ============================================
# 3. STAGED CHANGES (Commit Compliance)
# ============================================
echo "📝 STAGED CHANGES"
echo "------------------------------------------------"

if git diff --cached --quiet; then
    echo -e "${BLUE}No staged changes${NC}"
else
    FILE_COUNT=$(git diff --cached --name-only | wc -l | tr -d ' ')
    LINE_COUNT=$(git diff --cached --numstat | awk '{add+=$1; del+=$2} END {print add+del}')

    echo "Files staged: $FILE_COUNT"
    echo "Lines changed: $LINE_COUNT"

    # Check file count
    if [ "$FILE_COUNT" -gt 10 ]; then
        check_warn "More than 10 files staged (consider atomic commits)"
    fi

    # Check line count
    if [ "$LINE_COUNT" -gt 500 ]; then
        check_warn "More than 500 lines changed (consider splitting)"
    fi

    # Check for generated files
    GENERATED=$(git diff --cached --name-only | grep -E '\.(out|exe|dll|so|test)$' || true)
    if [ -n "$GENERATED" ]; then
        check_fail "Generated files in staging:"
        echo "$GENERATED" | sed 's/^/    /'
    fi

    # Check for coverage files
    COVERAGE_FILES=$(git diff --cached --name-only | grep -E 'coverage\.(out|html)' || true)
    if [ -n "$COVERAGE_FILES" ]; then
        check_fail "Coverage files in staging (should be in .gitignore):"
        echo "$COVERAGE_FILES" | sed 's/^/    /'
    fi

    # Check for debug statements
    DEBUG_PATTERNS='TODO|FIXME|XXX|fmt\.Println\("DEBUG'
    if git diff --cached | grep -E "$DEBUG_PATTERNS" > /dev/null 2>&1; then
        check_warn "Debug statements found in staged changes"
    fi
fi

echo ""

# ============================================

echo ""

# ============================================
# 5. DOCUMENTATION
# ============================================
echo "📚 DOCUMENTATION"
echo "------------------------------------------------"

# README
echo -n "README.md: "
if [ -f "README.md" ]; then
    LINES=$(wc -l < README.md)
    if [ "$LINES" -gt 10 ]; then
        check_pass
    else
        check_warn "README exists but is sparse (${LINES} lines)"
    fi
else
    check_fail "Missing README.md"
fi

# AGENTS.md
echo -n "AGENTS.md: "
if [ -f "AGENTS.md" ]; then
    check_pass
else
    check_warn "Consider creating AGENTS.md for handover documentation"
fi

# .gitignore
echo -n ".gitignore: "
if [ -f ".gitignore" ]; then
    if grep -q "coverage.out" .gitignore && grep -q "*.exe" .gitignore; then
        check_pass
    else
        check_warn "gitignore missing common patterns"
    fi
else
    check_fail "Missing .gitignore"
fi


echo ""

# ============================================
# 6. TESTING STANDARDS (Ginkgo/Gomega)
# ============================================
echo "🧪 TESTING STANDARDS"
echo "------------------------------------------------"

# Check for test files
TEST_COUNT=$(find . -name "*_test.go" -type f | wc -l)
GO_FILES=$(find . -name "*.go" -not -name "*_test.go" -not -path "./vendor/*" -type f | wc -l)

echo -n "Test Files: "
if [ "$TEST_COUNT" -gt 0 ]; then
    RATIO=$(awk "BEGIN {printf \"%.1f\", $TEST_COUNT/$GO_FILES}")
    echo -e "${GREEN}${TEST_COUNT} test files${NC} (ratio: ${RATIO}:1)"
else
    check_fail "No test files found"
fi

# Check for Ginkgo suite files
echo -n "Ginkgo Suites: "
SUITE_COUNT=$(find . -name "suite_test.go" -type f | wc -l)
if [ "$SUITE_COUNT" -gt 0 ]; then
    check_pass
else
    check_warn "No Ginkgo suite files found"
fi

# Check for skipped/pending tests (PROHIBITED)
echo -n "No Skipped/Pending Tests: "
# Find Skip() calls that are NOT environment-conditional (integration tests)
# Environment-conditional skips are allowed (clipboard, display, database path checks)
SKIPPED_TESTS=$(grep -rn "Skip(" --include="*_test.go" . 2>/dev/null | \
    grep -v "integration_test.go" | \
    grep -v "e2e_test.go" | \
    grep -v "Skipping integration test" | \
    grep -v "Database not found" | \
    grep -v "clipboard not supported" | \
    grep -v "no display available" | \
    grep -v "no clipboard utilities" || true)

if [ -z "$SKIPPED_TESTS" ]; then
    check_pass
else
    SKIP_COUNT=$(echo "$SKIPPED_TESTS" | wc -l)
    check_fail "Found $SKIP_COUNT skipped/pending tests (PROHIBITED)"
    echo ""
    echo -e "${RED}  Skipped/pending tests are NOT allowed before commits.${NC}"
    echo -e "${RED}  Either implement the tests or remove them.${NC}"
    echo ""
    echo "  Files with skipped tests:"
    echo "$SKIPPED_TESTS" | cut -d: -f1 | sort -u | sed 's/^/    /'
    echo ""
    echo "  To see all skipped tests: grep -rn 'Skip(' --include='*_test.go' ."
fi

echo ""

# ============================================
# 7. DEPENDENCY HEALTH
# ============================================
echo "📦 DEPENDENCY HEALTH"
echo "------------------------------------------------"

# go.mod exists
echo -n "go.mod: "
if [ -f "go.mod" ]; then
    check_pass
else
    check_fail "Missing go.mod"
fi

# go.sum exists
echo -n "go.sum: "
if [ -f "go.sum" ]; then
    check_pass
else
    check_warn "Missing go.sum (run: go mod tidy)"
fi

# Check for tidiness
echo -n "Module Tidiness: "
if go mod tidy -diff > /dev/null 2>&1; then
    check_pass
else
    check_warn "Modules not tidy (run: go mod tidy)"
fi

echo ""

# ============================================
# 8. FILE ORGANIZATION
# ============================================
echo "📁 FILE ORGANIZATION"
echo "------------------------------------------------"

# Check for proper internal/ structure
echo -n "Internal Structure: "
if [ -d "internal" ]; then
    check_pass
else
    check_warn "No internal/ directory (non-standard)"
fi

# Check for cmd/ structure
echo -n "Binary Structure: "
if [ -d "cmd" ]; then
    check_pass
else
    check_warn "No cmd/ directory (consider for binaries)"
fi

# Check for Makefile
echo -n "Makefile: "
if [ -f "Makefile" ]; then
    check_pass
else
    check_warn "No Makefile (consider for task automation)"
fi

echo ""

# ============================================
# 9. GIT HEALTH
# ============================================
echo "🔀 GIT HEALTH"
echo "------------------------------------------------"

# Check if in git repo
echo -n "Git Repository: "
if git rev-parse --git-dir > /dev/null 2>&1; then
    check_pass
else
    check_fail "Not a git repository"
fi

# Check for uncommitted changes (excluding staged)
echo -n "Working Directory: "
if git diff --quiet; then
    echo -e "${GREEN}Clean${NC}"
else
    echo -e "${YELLOW}Has unstaged changes${NC}"
fi

# Check recent commit message quality (if commits exist)
if git log -1 --pretty=%B > /dev/null 2>&1; then
    echo -n "Last Commit Format: "
    LAST_MSG=$(git log -1 --pretty=%B | head -1)
    if echo "$LAST_MSG" | grep -qE '^(feat|fix|docs|style|refactor|test|chore|perf)(\(.+\))?: .{1,50}$'; then
        check_pass
    else
        check_warn "Last commit doesn't follow conventional format"
    fi
fi

echo ""

echo ""

# ============================================
# SUMMARY
# ============================================
echo "================================================"
echo "SUMMARY"
echo "================================================"

if [ $VIOLATIONS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ ALL CHECKS PASSED${NC}"
    echo ""
    echo "Your project is compliant with all rules!"
    exit 0
elif [ $VIOLATIONS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  ${WARNINGS} WARNING(S) FOUND${NC}"
    echo ""
    echo "Project is functional but has minor issues."
    echo "Review warnings above for improvements."
    exit 0
else
    echo -e "${RED}❌ ${VIOLATIONS} VIOLATION(S) FOUND${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}⚠️  ${WARNINGS} WARNING(S) FOUND${NC}"
    fi
    echo ""
    echo "Fix violations before proceeding:"
    echo "  1. Review failed checks above"
    echo "  2. Apply suggested fixes"
    echo "  3. Re-run: make check-compliance"
    exit 1
fi

