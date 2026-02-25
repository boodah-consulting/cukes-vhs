#!/bin/bash

# Install Git Hooks for AI Commit Attribution
#
# Configures git to use .git-hooks/ as the hooks directory.
# This means hooks are tracked in the repo and always in sync.

set -e

echo "================================================"
echo "Installing Git Hooks"
echo "================================================"
echo ""

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ ! -e ".git" ]; then
    echo "Error: Not in a git repository"
    exit 1
fi

if [ ! -d ".git-hooks" ]; then
    echo "Error: .git-hooks/ directory not found"
    exit 1
fi

git config core.hooksPath .git-hooks
echo -e "${GREEN}Set core.hooksPath to .git-hooks/${NC}"
echo ""

REQUIRED_HOOKS=("pre-commit" "commit-msg" "prepare-commit-msg")
for hook in "${REQUIRED_HOOKS[@]}"; do
    hook_path=".git-hooks/$hook"
    if [ -f "$hook_path" ]; then
        chmod +x "$hook_path"
        echo -e "  ${GREEN}$hook${NC}"
    else
        echo -e "  ${YELLOW}$hook (missing from .git-hooks/)${NC}"
    fi
done

if [ -f "package.json" ] && command -v npm &> /dev/null; then
    echo ""
    echo "Installing Node.js dependencies for commitlint..."
    if npm ci 2>/dev/null || npm install 2>/dev/null; then
        echo -e "${GREEN}Installed Node.js dependencies${NC}"
    else
        echo -e "${YELLOW}Could not install Node.js dependencies${NC}"
    fi
fi

if git config commit.template .gitmessage 2>/dev/null; then
    echo -e "${GREEN}Configured commit template${NC}"
fi

echo ""
echo "================================================"
echo "Installation Complete"
echo "================================================"
echo ""
echo "Git now uses .git-hooks/ directly (tracked in repo)."
echo "Hooks stay in sync automatically - no manual copying needed."
echo ""
echo "Installed hooks:"
echo "  - pre-commit: Code quality and TDD enforcement"
echo "  - prepare-commit-msg: Adds AI attribution reminder"
echo "  - commit-msg: Validates AI attribution format"
echo ""
echo -e "${BLUE}To bypass hooks (emergencies only):${NC}"
echo "  git commit --no-verify"
echo ""
