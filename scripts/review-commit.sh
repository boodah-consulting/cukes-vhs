#!/usr/bin/env bash
# Pre-commit review script for cukes-vhs
# Non-interactive — reviews staged changes before committing
#
# Checks:
#   1. Staged files exist
#   2. Change statistics
#   3. Generated files check
#   4. Debug code detection
#   5. Secrets pattern detection
#   6. Build check
#   7. Test check
#   8. Format check
#   9. Architecture check (no import cycles)
#  10. AI attribution check

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
TOTAL_CHECKS=10

check()    { echo -e "\n${BOLD}${BLUE}[$1/${TOTAL_CHECKS}] $2${NC}"; }
pass()     { PASS=$((PASS + 1)); echo -e "  ${GREEN}✓${NC} $*"; }
fail()     { FAIL=$((FAIL + 1)); echo -e "  ${RED}✗${NC} $*"; }
warn_msg() { WARN=$((WARN + 1)); echo -e "  ${YELLOW}⚠${NC} $*"; }
skip_msg() { echo -e "  ${YELLOW}⊘${NC} $* (skipped)"; }

echo -e "${BOLD}cukes-vhs Pre-Commit Review${NC}"
echo "=============================="

START_TIME=$(date +%s)

# =============================================================================
# 1. Staged Files
# =============================================================================

check 1 "Staged Files"

STAGED_FILES=$(git diff --cached --name-only 2>/dev/null || true)

if [ -z "$STAGED_FILES" ]; then
  fail "No staged files found. Stage files with 'git add' first."
  echo ""
  echo -e "${RED}${BOLD}REVIEW ABORTED${NC} — nothing to review"
  exit 1
fi

STAGED_COUNT=$(echo "$STAGED_FILES" | wc -l | tr -d ' ')
pass "Found $STAGED_COUNT staged file(s)"
echo "$STAGED_FILES" | while IFS= read -r f; do echo "    $f"; done

# =============================================================================
# 2. Change Statistics
# =============================================================================

check 2 "Change Statistics"

STAT_OUTPUT=$(git diff --cached --stat 2>/dev/null || true)
INSERTIONS=$(echo "$STAT_OUTPUT" | tail -1 | grep -oE '[0-9]+ insertion' | grep -oE '[0-9]+' || echo "0")
DELETIONS=$(echo "$STAT_OUTPUT" | tail -1 | grep -oE '[0-9]+ deletion' | grep -oE '[0-9]+' || echo "0")

pass "Files changed: $STAGED_COUNT"
echo "    Insertions: +${INSERTIONS}"
echo "    Deletions:  -${DELETIONS}"

# Warn on large changes
TOTAL_CHANGES=$((INSERTIONS + DELETIONS))
if [ "$TOTAL_CHANGES" -gt 500 ]; then
  warn_msg "Large change set ($TOTAL_CHANGES lines) — consider splitting into smaller commits"
elif [ "$TOTAL_CHANGES" -gt 200 ]; then
  warn_msg "Moderate change set ($TOTAL_CHANGES lines)"
fi

# =============================================================================
# 3. Generated Files
# =============================================================================

check 3 "Generated Files"

GENERATED_PATTERNS=(
  "mock_*.go"
  "*_mock.go"
  "*.pb.go"
  "*_string.go"
  "*_gen.go"
  "generated_*.go"
)

GENERATED_FOUND=0
for pattern in "${GENERATED_PATTERNS[@]}"; do
  MATCHES=$(echo "$STAGED_FILES" | grep -E "$(echo "$pattern" | sed 's/\*/.*/g')" 2>/dev/null || true)
  if [ -n "$MATCHES" ]; then
    GENERATED_FOUND=$((GENERATED_FOUND + 1))
    warn_msg "Generated file staged: $MATCHES"
  fi
done

if [ "$GENERATED_FOUND" -eq 0 ]; then
  pass "No generated files in staged changes"
fi

# =============================================================================
# 4. Debug Code
# =============================================================================

check 4 "Debug Code Detection"

DEBUG_PATTERNS=(
  'fmt\.Println'
  'fmt\.Printf'
  'log\.Print'
  'log\.Println'
  'log\.Printf'
  'console\.log'
  'debugger'
  'TODO.*REMOVE'
  'FIXME.*REMOVE'
  'HACK'
)

DEBUG_FOUND=0
STAGED_GO_FILES=$(echo "$STAGED_FILES" | grep '\.go$' || true)

if [ -n "$STAGED_GO_FILES" ]; then
  for pattern in "${DEBUG_PATTERNS[@]}"; do
    # Check only staged content (not full file)
    MATCHES=$(git diff --cached -G "$pattern" --name-only 2>/dev/null || true)
    if [ -n "$MATCHES" ]; then
      # Verify the pattern is in added lines (not removed)
      ADDED_MATCHES=$(git diff --cached -U0 2>/dev/null | grep "^+" | grep -E "$pattern" 2>/dev/null || true)
      if [ -n "$ADDED_MATCHES" ]; then
        DEBUG_FOUND=$((DEBUG_FOUND + 1))
        warn_msg "Debug code pattern '${pattern}' found in staged additions:"
        echo "$ADDED_MATCHES" | head -3 | while IFS= read -r line; do echo "    $line"; done
      fi
    fi
  done
fi

if [ "$DEBUG_FOUND" -eq 0 ]; then
  pass "No debug code patterns found"
fi

# =============================================================================
# 5. Secrets Detection
# =============================================================================

check 5 "Secrets Pattern Detection"

SECRET_PATTERNS=(
  'AKIA[0-9A-Z]{16}'
  'password\s*[:=]\s*["\x27][^"\x27]+'
  'api[_-]?key\s*[:=]\s*["\x27][^"\x27]+'
  'secret[_-]?key\s*[:=]\s*["\x27][^"\x27]+'
  'private[_-]?key\s*[:=]\s*["\x27][^"\x27]+'
  'BEGIN RSA PRIVATE KEY'
  'BEGIN OPENSSH PRIVATE KEY'
)

SECRETS_FOUND=0
for pattern in "${SECRET_PATTERNS[@]}"; do
  # Check staged additions only
  ADDED_SECRETS=$(git diff --cached -U0 2>/dev/null | grep "^+" | grep -iE "$pattern" 2>/dev/null || true)
  if [ -n "$ADDED_SECRETS" ]; then
    SECRETS_FOUND=$((SECRETS_FOUND + 1))
    fail "Possible secret pattern detected: ${pattern}"
    echo "$ADDED_SECRETS" | head -2 | while IFS= read -r line; do echo "    $line"; done
  fi
done

# Check for .env files
ENV_STAGED=$(echo "$STAGED_FILES" | grep -E '\.env' || true)
if [ -n "$ENV_STAGED" ]; then
  SECRETS_FOUND=$((SECRETS_FOUND + 1))
  fail "Environment file staged: $ENV_STAGED"
fi

if [ "$SECRETS_FOUND" -eq 0 ]; then
  pass "No secret patterns detected"
fi

# =============================================================================
# 6. Build Check
# =============================================================================

check 6 "Build Check"

if go build ./... 2>&1; then
  pass "Build succeeded"
else
  fail "Build failed"
fi

# =============================================================================
# 7. Test Check
# =============================================================================

check 7 "Test Check"

if command -v ginkgo &>/dev/null; then
  if ginkgo --race --skip-package=testdata ./... 2>&1; then
    pass "All tests passed"
  else
    fail "Tests failed"
  fi
else
  skip_msg "ginkgo not installed"
fi

# =============================================================================
# 8. Format Check
# =============================================================================

check 8 "Format Check"

UNFORMATTED=""
if [ -n "$STAGED_GO_FILES" ]; then
  for f in $STAGED_GO_FILES; do
    if [ -f "$f" ]; then
      RESULT=$(gofmt -l "$f" 2>/dev/null || true)
      if [ -n "$RESULT" ]; then
        UNFORMATTED="${UNFORMATTED}${RESULT}\n"
      fi
    fi
  done
fi

if [ -z "$UNFORMATTED" ]; then
  pass "All staged Go files properly formatted"
else
  fail "Unformatted files:"
  echo -e "$UNFORMATTED" | while IFS= read -r f; do [ -n "$f" ] && echo "    $f"; done
fi

# =============================================================================
# 9. Architecture Check
# =============================================================================

check 9 "Architecture Check"

# Verify no import cycles
if go vet ./... 2>/dev/null; then
  pass "No import cycles detected (go vet passed)"
else
  fail "go vet failed — possible import cycles"
fi

# Check dependency direction: internal/ should not import cmd/
CMD_IMPORTS=$(echo "$STAGED_GO_FILES" | while IFS= read -r f; do
  if echo "$f" | grep -q '^internal/'; then
    if [ -f "$f" ]; then
      grep '"github.com/boodah-consulting/cukes-vhs/cmd' "$f" 2>/dev/null || true
    fi
  fi
done)

if [ -z "$CMD_IMPORTS" ]; then
  pass "Correct dependency direction (internal/ does not import cmd/)"
else
  fail "internal/ imports cmd/ (wrong dependency direction)"
fi

# =============================================================================
# 10. AI Attribution Check
# =============================================================================

check 10 "AI Attribution"

# Check if running in an AI agent context
if [ -n "${Claude:-}" ] || [ -n "${CLAUDE_CODE:-}" ] || [ -n "${OPENCODE:-}" ] || [ -n "${CURSOR:-}" ] || [ -n "${AI_AGENT:-}" ]; then
  warn_msg "AI agent detected — ensure commit uses 'make ai-commit' for proper attribution"
else
  pass "No AI agent context detected (manual commit)"
fi

# =============================================================================
# Summary
# =============================================================================

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo -e "${BOLD}=============================="
echo -e "Review Summary${NC}"
echo "=============================="
echo "  Duration: ${DURATION}s"
echo -e "  ${GREEN}Passed:${NC}   $PASS"
echo -e "  ${RED}Failed:${NC}   $FAIL"
echo -e "  ${YELLOW}Warnings:${NC} $WARN"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo -e "${RED}${BOLD}REVIEW FAILED${NC} — $FAIL issue(s) must be resolved before committing"
  exit 1
else
  if [ "$WARN" -gt 0 ]; then
    echo -e "${YELLOW}${BOLD}REVIEW PASSED WITH WARNINGS${NC} — $WARN warning(s) to consider"
  else
    echo -e "${GREEN}${BOLD}REVIEW PASSED${NC} — ready to commit"
  fi
  exit 0
fi
