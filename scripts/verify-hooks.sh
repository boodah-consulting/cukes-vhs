#!/bin/bash

# Verify Git Hooks Installation Script
#
# Checks that core.hooksPath is set to .git-hooks/ and all required hooks exist.

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REQUIRED_HOOKS=("pre-commit" "commit-msg" "prepare-commit-msg")
MISSING_HOOKS=()
INVALID_HOOKS=()

echo "Git Hooks Verification"
echo ""

if [ ! -e ".git" ]; then
    echo -e "${RED}Error: Not in a git repository${NC}"
    exit 1
fi

HOOKS_PATH=$(git config --get core.hooksPath 2>/dev/null || true)

if [ "$HOOKS_PATH" != ".git-hooks" ]; then
    echo -e "${YELLOW}core.hooksPath is not set to .git-hooks/${NC}"
    echo ""
    echo -e "${BLUE}Run: make install-git-hooks${NC}"
    echo ""
    exit 1
fi

echo -e "core.hooksPath: ${GREEN}.git-hooks/${NC}"
echo ""

for hook in "${REQUIRED_HOOKS[@]}"; do
    hook_path=".git-hooks/$hook"

    echo -n "  $hook: "

    if [ ! -f "$hook_path" ]; then
        echo -e "${RED}missing${NC}"
        MISSING_HOOKS+=("$hook")
    elif [ ! -x "$hook_path" ]; then
        echo -e "${YELLOW}not executable${NC}"
        INVALID_HOOKS+=("$hook")
    elif head -1 "$hook_path" | grep -q '^#!/'; then
        echo -e "${GREEN}ok${NC}"
    else
        echo -e "${YELLOW}invalid format${NC}"
        INVALID_HOOKS+=("$hook")
    fi
done

echo ""

if [ ${#MISSING_HOOKS[@]} -eq 0 ] && [ ${#INVALID_HOOKS[@]} -eq 0 ]; then
    echo -e "${GREEN}All git hooks are properly installed${NC}"
    exit 0
else
    echo -e "${RED}Git hooks installation incomplete${NC}"
    echo ""

    if [ ${#MISSING_HOOKS[@]} -gt 0 ]; then
        echo "Missing hooks:"
        for hook in "${MISSING_HOOKS[@]}"; do
            echo "  - $hook"
        done
        echo ""
    fi

    if [ ${#INVALID_HOOKS[@]} -gt 0 ]; then
        echo "Invalid/non-executable hooks:"
        for hook in "${INVALID_HOOKS[@]}"; do
            echo "  - $hook"
        done
        echo ""
    fi

    echo -e "${BLUE}Run: make install-git-hooks${NC}"
    echo ""
    exit 1
fi
