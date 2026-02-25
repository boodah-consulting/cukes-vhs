#!/bin/bash

# Pre-task validation for AI agents
# Run this BEFORE starting any task to ensure you're set up correctly

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "================================================"
echo "PRE-TASK CHECKLIST"
echo "================================================"
echo ""

READY=true

# Check 1: Session started?
echo -n "1. Session started (hooks verified): "
if [ -f ".git/hooks/pre-commit" ] && [ -f ".git/hooks/commit-msg" ]; then
    echo -e "${GREEN}YES${NC}"
else
    echo -e "${RED}NO${NC} - Run: make session-start"
    READY=false
fi

# Check 2: On correct branch?
echo -n "2. On feature branch (not main/master): "
BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ] && [ "$BRANCH" != "master" ]; then
    echo -e "${GREEN}YES${NC} ($BRANCH)"
else
    echo -e "${YELLOW}WARNING${NC} - On $BRANCH, consider creating feature branch"
fi

# Check 3: Working directory clean?
echo -n "3. Working directory clean: "
if git diff --quiet && git diff --cached --quiet; then
    echo -e "${GREEN}YES${NC}"
else
    echo -e "${YELLOW}NO${NC} - Uncommitted changes exist"
fi

# Check 4: Build passes?
echo -n "4. Build passes: "
if go build ./... 2>/dev/null; then
    echo -e "${GREEN}YES${NC}"
else
    echo -e "${RED}NO${NC} - Fix build errors first"
    READY=false
fi

# Check 5: Tests pass?
echo -n "5. Tests pass: "
if go test ./... >/dev/null 2>&1; then
    echo -e "${GREEN}YES${NC}"
else
    echo -e "${RED}NO${NC} - Fix failing tests first"
    READY=false
fi

# Check 6: Pattern check
echo -n "6. Pattern check: "
VIOLATIONS=0
DIRECT_HUH=$(grep -rn "form \*huh\.Form" internal/cli/intents/*.go 2>/dev/null | grep -v "_test.go" || true)
if [ -n "$DIRECT_HUH" ]; then VIOLATIONS=$((VIOLATIONS+1)); fi
HARDCODED=$(grep -rn "lipgloss\.Color(\"#[0-9A-Fa-f]" internal/cli/intents/*.go internal/cli/models/*.go 2>/dev/null | grep -v "_test.go" || true)
if [ -n "$HARDCODED" ]; then VIOLATIONS=$((VIOLATIONS+1)); fi

if [ $VIOLATIONS -eq 0 ]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${YELLOW}$VIOLATIONS issue(s)${NC} - Pre-existing, be aware"
fi

echo ""
echo "================================================"

if [ "$READY" = true ]; then
    echo -e "${GREEN}READY TO START TASK${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Identify the task from tasks/ directory"
    echo "  2. Write the TEST first (TDD red phase)"
    echo "  3. Run test - confirm it FAILS"
    echo "  4. Write minimal implementation"
    echo "  5. Run test - confirm it PASSES"
    echo "  6. Run: make check-compliance"
    echo "  7. Run: make ai-commit MSG=\"type(scope): description\""
else
    echo -e "${RED}NOT READY${NC}"
    echo ""
    echo "Fix the issues above before starting."
    exit 1
fi

echo "================================================"
