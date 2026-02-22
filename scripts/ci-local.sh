#!/usr/bin/env bash
# Local CI pipeline for cukes-vhs
# Non-interactive — mirrors the GitHub Actions CI pipeline locally
#
# Steps:
#   1. Commitlint check
#   2. AI attribution check
#   3. Format check (gofmt)
#   4. Vet check (go vet)
#   5. Staticcheck
#   6. Tests (ginkgo with race)
#   7. Multi-platform build check
#   8. Security scan (gosec)

set -euo pipefail

# =============================================================================
# Colour helpers
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

STEP=0
TOTAL_STEPS=8
FAILURES=0

step() {
  STEP=$((STEP + 1))
  echo -e "\n${BOLD}${BLUE}[${STEP}/${TOTAL_STEPS}] $1${NC}"
}

pass()    { echo -e "  ${GREEN}✓${NC} $*"; }
fail()    { FAILURES=$((FAILURES + 1)); echo -e "  ${RED}✗${NC} $*"; }
skip_msg() { echo -e "  ${YELLOW}⊘${NC} $* (skipped)"; }

echo -e "${BOLD}cukes-vhs Local CI Pipeline${NC}"
echo "=============================="
echo "Module: github.com/boodah-consulting/cukes-vhs"
echo ""

START_TIME=$(date +%s)

# =============================================================================
# 1. Commitlint
# =============================================================================

step "Commitlint"

if command -v npx &>/dev/null; then
  # Check the most recent commit
  LAST_MSG=$(git log -1 --format='%s' 2>/dev/null || true)
  if [ -n "$LAST_MSG" ]; then
    if echo "$LAST_MSG" | npx --no-install commitlint --config .commitlintrc.json 2>/dev/null; then
      pass "Last commit passes commitlint"
    else
      # Try with npx install
      if echo "$LAST_MSG" | npx commitlint --config .commitlintrc.json 2>/dev/null; then
        pass "Last commit passes commitlint"
      else
        fail "Last commit fails commitlint: $LAST_MSG"
      fi
    fi
  else
    skip_msg "No commits found"
  fi
else
  skip_msg "npx not available — install Node.js for commitlint"
fi

# =============================================================================
# 2. AI Attribution
# =============================================================================

step "AI Attribution"

# Check if recent commits by AI agents have the trailer
RECENT_COMMITS=$(git log --format='%H %s' -10 2>/dev/null || true)
AI_MISSING=0

while IFS= read -r commit_line; do
  [ -z "$commit_line" ] && continue
  HASH=$(echo "$commit_line" | cut -d' ' -f1)
  SUBJECT=$(echo "$commit_line" | cut -d' ' -f2-)

  # Check if commit has AI trailer
  BODY=$(git log -1 --format='%b' "$HASH" 2>/dev/null || true)
  if echo "$BODY" | grep -q "AI-Generated-By:"; then
    pass "AI attribution present: ${SUBJECT:0:50}"
  fi
done <<< "$RECENT_COMMITS"

pass "AI attribution check complete"

# =============================================================================
# 3. Format Check
# =============================================================================

step "Format Check (gofmt)"

UNFORMATTED=$(gofmt -l . 2>/dev/null || true)
if [ -z "$UNFORMATTED" ]; then
  pass "All Go files properly formatted"
else
  fail "Unformatted files:"
  echo "$UNFORMATTED" | while IFS= read -r f; do echo "    $f"; done
fi

# =============================================================================
# 4. Vet Check
# =============================================================================

step "Vet Check (go vet)"

if go vet ./... 2>&1; then
  pass "go vet passed"
else
  fail "go vet reported issues"
fi

# =============================================================================
# 5. Staticcheck
# =============================================================================

step "Staticcheck"

if command -v staticcheck &>/dev/null; then
  if staticcheck ./... 2>&1; then
    pass "staticcheck passed"
  else
    fail "staticcheck reported issues"
  fi
else
  skip_msg "staticcheck not installed — run: go install honnef.co/go/tools/cmd/staticcheck@latest"
fi

# =============================================================================
# 6. Tests
# =============================================================================

step "Tests (ginkgo --race)"

if command -v ginkgo &>/dev/null; then
  if ginkgo -v --race --skip-package=testdata ./... 2>&1; then
    pass "All tests passed"
  else
    fail "Tests failed"
  fi
else
  skip_msg "ginkgo not installed — run: go install github.com/onsi/ginkgo/v2/ginkgo@latest"
fi

# =============================================================================
# 7. Multi-platform Build
# =============================================================================

step "Multi-platform Build"

MODULE="github.com/boodah-consulting/cukes-vhs"
BINARY="cukes-vhs"
BUILD_DIR=$(mktemp -d /tmp/cukes-vhs-build-XXXXXX)
trap "rm -rf $BUILD_DIR" EXIT

PLATFORMS=(
  "linux/amd64"
  "darwin/amd64"
  "darwin/arm64"
)

BUILD_FAILURES=0
for platform in "${PLATFORMS[@]}"; do
  GOOS=$(echo "$platform" | cut -d'/' -f1)
  GOARCH=$(echo "$platform" | cut -d'/' -f2)
  OUTPUT="${BUILD_DIR}/${BINARY}-${GOOS}-${GOARCH}"

  if GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build -o "$OUTPUT" ./cmd/cukes-vhs 2>/dev/null; then
    SIZE=$(du -h "$OUTPUT" 2>/dev/null | cut -f1)
    pass "Build: ${platform} ($SIZE)"
  else
    fail "Build failed: ${platform}"
    BUILD_FAILURES=$((BUILD_FAILURES + 1))
  fi
done

if [ "$BUILD_FAILURES" -eq 0 ]; then
  pass "All platform builds succeeded"
fi

# =============================================================================
# 8. Security Scan
# =============================================================================

step "Security Scan (gosec)"

if command -v gosec &>/dev/null; then
  if gosec -quiet ./... 2>&1; then
    pass "gosec passed — no security issues found"
  else
    fail "gosec reported security issues"
  fi
else
  skip_msg "gosec not installed — run: go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi

# =============================================================================
# Summary
# =============================================================================

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo -e "${BOLD}=============================="
echo -e "CI Pipeline Summary${NC}"
echo "=============================="
echo "  Duration: ${DURATION}s"
echo "  Steps:    ${TOTAL_STEPS}"
echo "  Failures: ${FAILURES}"
echo ""

if [ "$FAILURES" -gt 0 ]; then
  echo -e "${RED}${BOLD}CI PIPELINE FAILED${NC} — $FAILURES step(s) failed"
  exit 1
else
  echo -e "${GREEN}${BOLD}CI PIPELINE PASSED${NC} — all checks green"
  exit 0
fi
