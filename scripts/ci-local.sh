#!/bin/bash
# ci-local.sh - Run all CI checks locally
# This script mirrors all checks from .github/workflows/ci.yml and pr-validation.yml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track failures
FAILED_CHECKS=()
PASSED_CHECKS=()

# Helper function to run a check
run_check() {
    local name="$1"
    local command="$2"
    
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Running: $name${NC}"
    echo -e "${BLUE}========================================${NC}"
    
    if eval "$command"; then
        echo -e "${GREEN}✅ PASSED: $name${NC}"
        PASSED_CHECKS+=("$name")
    else
        echo -e "${RED}❌ FAILED: $name${NC}"
        FAILED_CHECKS+=("$name")
        return 1
    fi
}

# Print header
echo -e "${BLUE}"
echo "================================================"
echo "         KaRiya - Local CI Checks"
echo "================================================"
echo -e "${NC}"
echo "This script runs all checks from CI locally."
echo ""

# Install required tools if missing
echo -e "${YELLOW}Checking required tools...${NC}"

# Check and install ginkgo
if ! command -v ginkgo &> /dev/null; then
    echo "Installing ginkgo..."
    go install github.com/onsi/ginkgo/v2/ginkgo@latest
fi

# Check and install staticcheck
if ! command -v staticcheck &> /dev/null; then
    echo "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

# Check and install gosec
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Check and install golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.8
fi

# Check npm dependencies
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm ci
fi

echo -e "${GREEN}All required tools are installed.${NC}"
echo ""

# ============================================
# 1. COMMITLINT (from pr-validation.yml)
# ============================================
# Note: Only runs on last commit, not full PR range
if git rev-parse --git-dir > /dev/null 2>&1; then
    run_check "Commitlint (last commit)" \
        "git log -1 --pretty=%B | npx commitlint --config .commitlintrc.json" || true
fi

# ============================================
# 2. AI ATTRIBUTION CHECK (from pr-validation.yml)
# ============================================
if git rev-parse --git-dir > /dev/null 2>&1; then
    run_check "AI Attribution (last commit)" \
        "git log -1 --pretty=%B | grep -q 'AI-Generated-By:' || echo 'ℹ️  No AI attribution (manual commit or non-code)'" || true
fi

# ============================================
# 3. LINT & FORMAT (from ci.yml - lint job)
# ============================================
run_check "go fmt" \
    "test -z \"\$(gofmt -l .)\" || (echo 'Files need formatting:'; gofmt -l .; exit 1)"

run_check "go vet" \
    "go vet ./..."

run_check "staticcheck" \
    "staticcheck ./..."

# ============================================
# 4. TESTS (from ci.yml - test job)
# ============================================
run_check "Tests with race detector and coverage" \
    "ginkgo -v --race --cover --coverprofile=coverage.out --skip-package=testdata,noinlinecareer/testdata,features ./..."

# ============================================
# 5. BUILD (from ci.yml - build job)
# ============================================
run_check "Build Linux AMD64" \
    "GOOS=linux GOARCH=amd64 go build -o cukes-vhs-linux-amd64 ./cmd/cukes-vhs && chmod +x cukes-vhs-linux-amd64"

run_check "Build macOS AMD64" \
    "GOOS=darwin GOARCH=amd64 go build -o cukes-vhs-darwin-amd64 ./cmd/cukes-vhs && chmod +x cukes-vhs-darwin-amd64"

run_check "Build macOS ARM64" \
    "GOOS=darwin GOARCH=arm64 go build -o cukes-vhs-darwin-arm64 ./cmd/cukes-vhs && chmod +x cukes-vhs-darwin-arm64"

run_check "Build Windows AMD64" \
    "GOOS=windows GOARCH=amd64 go build -o cukes-vhs-windows-amd64.exe ./cmd/cukes-vhs"

# ============================================
# 6. SECURITY SCAN (from ci.yml - security job)
# ============================================
run_check "Gosec Security Scanner" \
    "gosec -no-fail -fmt text ./..."

# ============================================
# 7. GOLANGCI-LINT (comprehensive static analysis)
# ============================================
run_check "Golangci-lint" \
    "golangci-lint run --timeout=5m"

# ============================================
# 8. DOCBLOCKS (structured doc comment enforcement)
# ============================================
run_check "Docblocks Analyzer" \
    "go build -o ./bin/docblocks ./cmd/docblocks && go vet -vettool=./bin/docblocks ./internal/cli/behaviors/... ./internal/cli/intents/... ./tools/analyzers/docblocks/..."

# ============================================
# SUMMARY
# ============================================
echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}              CI CHECK SUMMARY${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

if [ ${#PASSED_CHECKS[@]} -gt 0 ]; then
    echo -e "${GREEN}✅ Passed Checks (${#PASSED_CHECKS[@]}):${NC}"
    for check in "${PASSED_CHECKS[@]}"; do
        echo -e "  ${GREEN}✓${NC} $check"
    done
    echo ""
fi

if [ ${#FAILED_CHECKS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Failed Checks (${#FAILED_CHECKS[@]}):${NC}"
    for check in "${FAILED_CHECKS[@]}"; do
        echo -e "  ${RED}✗${NC} $check"
    done
    echo ""
    echo -e "${RED}Some checks failed. Please fix the issues above.${NC}"
    exit 1
else
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}    ALL CI CHECKS PASSED! ✅${NC}"
    echo -e "${GREEN}================================================${NC}"
    echo ""
    echo "Your code is ready for CI. All checks that run in GitHub"
    echo "Actions have passed locally."
    echo ""
fi

# Clean up build artifacts
echo "Cleaning up build artifacts..."
rm -f cukes-vhs-linux-amd64 cukes-vhs-darwin-amd64 cukes-vhs-darwin-arm64 cukes-vhs-windows-amd64.exe

echo "Done!"
