#!/usr/bin/env bash
# Compliance checks for cukes-vhs
# Non-interactive — suitable for CI and pre-commit workflows
#
# Sections:
#   1. Code quality (fmt, vet, staticcheck)
#   2. Test coverage (ginkgo with threshold check)
#   3. Staged changes check
#   4. Architecture (verify internal/ structure, no circular imports)
#   5. Documentation (doc.go for packages)
#   6. Testing standards (test files exist)
#   7. Dependency health (go mod verify/tidy)
#   8. File organisation (expected directories)
#   9. Git health (no large files, no secrets)

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

PASS=0
FAIL=0
WARN=0

COVERAGE_THRESHOLD=90

section()  { echo -e "\n${BOLD}${BLUE}[$1/${TOTAL_SECTIONS}] $2${NC}"; }
pass()     { PASS=$((PASS + 1)); echo -e "  ${GREEN}✓${NC} $*"; }
fail()     { FAIL=$((FAIL + 1)); echo -e "  ${RED}✗${NC} $*"; }
skip()     { echo -e "  ${YELLOW}⊘${NC} $* (skipped)"; }
warn_msg() { WARN=$((WARN + 1)); echo -e "  ${YELLOW}⚠${NC} $*"; }

TOTAL_SECTIONS=9

echo -e "${BOLD}cukes-vhs Compliance Check${NC}"
echo "=============================="

# =============================================================================
# 1. Code Quality
# =============================================================================

section 1 "Code Quality"

# gofmt
UNFORMATTED=$(gofmt -l . 2>/dev/null || true)
if [ -z "$UNFORMATTED" ]; then
  pass "All files formatted (gofmt)"
else
  fail "Unformatted files found:"
  echo "$UNFORMATTED" | while IFS= read -r f; do echo "    $f"; done
fi

# go vet
if go vet ./... 2>/dev/null; then
  pass "go vet passed"
else
  fail "go vet reported issues"
fi

# staticcheck
if command -v staticcheck &>/dev/null; then
  if staticcheck ./... 2>/dev/null; then
    pass "staticcheck passed"
  else
    fail "staticcheck reported issues"
  fi
else
  skip "staticcheck not installed"
fi

# =============================================================================
# 2. Test Coverage
# =============================================================================

section 2 "Test Coverage"

if command -v ginkgo &>/dev/null; then
  COVERAGE_FILE=$(mktemp coverage-XXXXXX.out)
  trap "rm -f $COVERAGE_FILE" EXIT

  if ginkgo --race --skip-package=testdata --coverprofile="$COVERAGE_FILE" --covermode=atomic ./... 2>/dev/null; then
    pass "All tests passed"

    # Extract total coverage percentage
    if [ -s "$COVERAGE_FILE" ]; then
      COVERAGE_LINE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | tail -1 || true)
      COVERAGE_PCT=$(echo "$COVERAGE_LINE" | grep -oE '[0-9]+\.[0-9]+' | tail -1 || echo "0")
      COVERAGE_INT=${COVERAGE_PCT%.*}

      if [ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ]; then
        pass "Coverage: ${COVERAGE_PCT}% (threshold: ${COVERAGE_THRESHOLD}%)"
      else
        fail "Coverage: ${COVERAGE_PCT}% (below threshold: ${COVERAGE_THRESHOLD}%)"
      fi
    else
      warn_msg "Coverage file empty — no coverage data collected"
    fi
  else
    fail "Tests failed"
  fi
else
  skip "ginkgo not installed"
fi

# =============================================================================
# 3. Staged Changes
# =============================================================================

section 3 "Staged Changes"

STAGED=$(git diff --cached --name-only 2>/dev/null || true)
if [ -n "$STAGED" ]; then
  STAGED_COUNT=$(echo "$STAGED" | wc -l | tr -d ' ')
  warn_msg "Found $STAGED_COUNT staged file(s) — ensure they are intentional"
  echo "$STAGED" | while IFS= read -r f; do echo "    $f"; done
else
  pass "No staged changes (clean working state)"
fi

# =============================================================================
# 4. Architecture
# =============================================================================

section 4 "Architecture"

# Verify internal/ structure
if [ -d "internal/vhsgen" ]; then
  pass "internal/vhsgen/ exists"
else
  fail "internal/vhsgen/ not found"
fi

# Check cmd directory
if [ -d "cmd/cukes-vhs" ]; then
  pass "cmd/cukes-vhs/ exists"
else
  fail "cmd/cukes-vhs/ not found"
fi

# Check for circular imports
if go build ./... 2>/dev/null; then
  pass "No circular imports (build succeeded)"
else
  fail "Build failed — possible circular imports or compilation errors"
fi

# Verify no direct imports from cmd in internal
CMD_IN_INTERNAL=$(grep -r '"github.com/boodah-consulting/cukes-vhs/cmd' internal/ 2>/dev/null || true)
if [ -z "$CMD_IN_INTERNAL" ]; then
  pass "No cmd/ imports in internal/ (correct dependency direction)"
else
  fail "internal/ imports cmd/ (wrong dependency direction)"
  echo "$CMD_IN_INTERNAL" | while IFS= read -r line; do echo "    $line"; done
fi

# =============================================================================
# 5. Documentation
# =============================================================================

section 5 "Documentation"

# Check for doc.go in key packages
DOC_MISSING=0
for pkg_dir in internal/vhsgen cmd/cukes-vhs; do
  if [ -d "$pkg_dir" ]; then
    if [ -f "$pkg_dir/doc.go" ]; then
      pass "doc.go exists in $pkg_dir"
    else
      warn_msg "No doc.go in $pkg_dir"
      DOC_MISSING=$((DOC_MISSING + 1))
    fi
  fi
done

# Check sub-packages in internal/vhsgen
if [ -d "internal/vhsgen" ]; then
  for sub_dir in internal/vhsgen/*/; do
    if [ -d "$sub_dir" ]; then
      pkg_name=$(basename "$sub_dir")
      # Skip testdata directories
      if [ "$pkg_name" = "testdata" ]; then
        continue
      fi
      if [ -f "${sub_dir}doc.go" ]; then
        pass "doc.go exists in $sub_dir"
      else
        warn_msg "No doc.go in $sub_dir"
        DOC_MISSING=$((DOC_MISSING + 1))
      fi
    fi
  done
fi

# =============================================================================
# 6. Testing Standards
# =============================================================================

section 6 "Testing Standards"

# Check test files exist alongside source files
TEST_MISSING=0
for go_file in $(find internal/ -name '*.go' -not -name '*_test.go' -not -name 'doc.go' -not -path '*/testdata/*' 2>/dev/null || true); do
  dir=$(dirname "$go_file")
  base=$(basename "$go_file" .go)
  test_file="${dir}/${base}_test.go"
  suite_file="${dir}/${base}_suite_test.go"

  # Check for any test file in the same directory
  if ls "${dir}"/*_test.go 1>/dev/null 2>&1; then
    continue
  else
    warn_msg "No test files in directory: $dir"
    TEST_MISSING=$((TEST_MISSING + 1))
  fi
done

if [ "$TEST_MISSING" -eq 0 ]; then
  pass "All source directories have test files"
fi

# Verify Ginkgo suite files
SUITE_MISSING=0
for pkg_dir in $(find internal/ -type d -not -path '*/testdata/*' 2>/dev/null || true); do
  # Check if directory has Go files
  if ls "${pkg_dir}"/*.go 1>/dev/null 2>&1; then
    # Check if it has test files
    if ls "${pkg_dir}"/*_test.go 1>/dev/null 2>&1; then
      # Check for suite file
      if ! ls "${pkg_dir}"/*_suite_test.go 1>/dev/null 2>&1; then
        warn_msg "No Ginkgo suite file in: $pkg_dir"
        SUITE_MISSING=$((SUITE_MISSING + 1))
      fi
    fi
  fi
done

if [ "$SUITE_MISSING" -eq 0 ]; then
  pass "All test packages have Ginkgo suite files"
fi

# =============================================================================
# 7. Dependency Health
# =============================================================================

section 7 "Dependency Health"

# go mod verify
if go mod verify 2>/dev/null; then
  pass "go mod verify passed"
else
  fail "go mod verify failed — modules may be corrupted"
fi

# go mod tidy check (verify no changes needed)
TIDY_BEFORE=$(cat go.sum)
go mod tidy 2>/dev/null
TIDY_AFTER=$(cat go.sum)

if [ "$TIDY_BEFORE" = "$TIDY_AFTER" ]; then
  pass "go mod tidy — no changes needed"
else
  warn_msg "go mod tidy made changes — run 'go mod tidy' and commit"
  # Restore original state
  git checkout -- go.sum go.mod 2>/dev/null || true
fi

# =============================================================================
# 8. File Organisation
# =============================================================================

section 8 "File Organisation"

EXPECTED_DIRS=("cmd/cukes-vhs" "internal/vhsgen" "scripts")

for dir in "${EXPECTED_DIRS[@]}"; do
  if [ -d "$dir" ]; then
    pass "Directory exists: $dir"
  else
    fail "Expected directory missing: $dir"
  fi
done

# Check for unexpected top-level Go files (should be in cmd/ or internal/)
TOP_LEVEL_GO=$(find . -maxdepth 1 -name '*.go' -not -name '*_test.go' 2>/dev/null || true)
if [ -z "$TOP_LEVEL_GO" ]; then
  pass "No stray top-level Go files"
else
  warn_msg "Found top-level Go files (should be in cmd/ or internal/):"
  echo "$TOP_LEVEL_GO" | while IFS= read -r f; do echo "    $f"; done
fi

# =============================================================================
# 9. Git Health
# =============================================================================

section 9 "Git Health"

# Check for large files (>1MB)
LARGE_FILES=$(find . -type f -not -path './.git/*' -not -path './vendor/*' -size +1M 2>/dev/null || true)
if [ -z "$LARGE_FILES" ]; then
  pass "No large files (>1MB)"
else
  warn_msg "Large files found (>1MB):"
  echo "$LARGE_FILES" | while IFS= read -r f; do
    SIZE=$(du -h "$f" 2>/dev/null | cut -f1)
    echo "    $f ($SIZE)"
  done
fi

# Check for secrets patterns
SECRETS_FOUND=0
SECRET_PATTERNS=(
  'AKIA[0-9A-Z]{16}'           # AWS Access Key
  'password\s*=\s*["\x27][^"\x27]+'  # Hardcoded passwords
  'api[_-]?key\s*=\s*["\x27][^"\x27]+'  # API keys
  'secret[_-]?key\s*=\s*["\x27][^"\x27]+'  # Secret keys
  'token\s*=\s*["\x27][A-Za-z0-9+/=]{20,}'  # Tokens
)

for pattern in "${SECRET_PATTERNS[@]}"; do
  MATCHES=$(grep -rlE "$pattern" --include='*.go' --include='*.yaml' --include='*.yml' --include='*.json' --include='*.toml' --include='*.env' . 2>/dev/null | grep -v '.git/' | grep -v 'vendor/' || true)
  if [ -n "$MATCHES" ]; then
    SECRETS_FOUND=$((SECRETS_FOUND + 1))
    warn_msg "Possible secret pattern found ($pattern):"
    echo "$MATCHES" | while IFS= read -r f; do echo "    $f"; done
  fi
done

if [ "$SECRETS_FOUND" -eq 0 ]; then
  pass "No secret patterns detected"
fi

# Check for .env files committed
ENV_FILES=$(git ls-files '*.env' '.env*' 2>/dev/null || true)
if [ -z "$ENV_FILES" ]; then
  pass "No .env files tracked"
else
  fail "Found tracked .env files (should be in .gitignore):"
  echo "$ENV_FILES" | while IFS= read -r f; do echo "    $f"; done
fi

# =============================================================================
# Summary
# =============================================================================

echo ""
echo -e "${BOLD}=============================="
echo -e "Compliance Summary${NC}"
echo "=============================="
echo -e "  ${GREEN}Passed:${NC}   $PASS"
echo -e "  ${RED}Failed:${NC}   $FAIL"
echo -e "  ${YELLOW}Warnings:${NC} $WARN"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo -e "${RED}${BOLD}COMPLIANCE CHECK FAILED${NC} — $FAIL issue(s) must be resolved"
  exit 1
else
  if [ "$WARN" -gt 0 ]; then
    echo -e "${YELLOW}${BOLD}COMPLIANCE PASSED WITH WARNINGS${NC} — $WARN warning(s)"
  else
    echo -e "${GREEN}${BOLD}ALL COMPLIANCE CHECKS PASSED${NC}"
  fi
  exit 0
fi
