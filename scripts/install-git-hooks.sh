#!/usr/bin/env bash
# Install git hooks for cukes-vhs
# Non-interactive — auto-installs without prompts
#
# This script:
#   1. Verifies we are in a git repository
#   2. Configures git to use the .git-hooks directory
#   3. Ensures all hook files are executable
#   4. Verifies installation

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

info()    { echo -e "${BLUE}ℹ${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn()    { echo -e "${YELLOW}⚠${NC} $*"; }
error()   { echo -e "${RED}✗${NC} $*" >&2; }

echo -e "${BOLD}cukes-vhs Git Hooks Installer${NC}"
echo "================================"

# =============================================================================
# Pre-flight checks
# =============================================================================

# Verify we are in a git repository
if ! git rev-parse --git-dir >/dev/null 2>&1; then
  error "Not a git repository. Run this from the project root."
  exit 1
fi

# Get the project root (where .git-hooks should be)
PROJECT_ROOT=$(git rev-parse --show-toplevel)
HOOKS_DIR="${PROJECT_ROOT}/.git-hooks"

# Verify hooks directory exists
if [ ! -d "$HOOKS_DIR" ]; then
  error "Hooks directory not found: $HOOKS_DIR"
  error "Expected .git-hooks/ directory at project root"
  exit 1
fi

info "Project root: $PROJECT_ROOT"
info "Hooks source: $HOOKS_DIR"

# =============================================================================
# Configure git hooks path
# =============================================================================

CURRENT_HOOKS_PATH=$(git config --get core.hooksPath 2>/dev/null || true)

if [ "$CURRENT_HOOKS_PATH" = ".git-hooks" ]; then
  info "Git hooks path already configured to .git-hooks"
else
  git config core.hooksPath .git-hooks
  success "Git hooks path set to .git-hooks"
fi

# =============================================================================
# Set executable permissions
# =============================================================================

HOOK_COUNT=0
for hook_file in "$HOOKS_DIR"/*; do
  if [ -f "$hook_file" ]; then
    chmod +x "$hook_file"
    HOOK_NAME=$(basename "$hook_file")
    success "Made executable: $HOOK_NAME"
    HOOK_COUNT=$((HOOK_COUNT + 1))
  fi
done

if [ "$HOOK_COUNT" -eq 0 ]; then
  warn "No hook files found in $HOOKS_DIR"
  exit 0
fi

# =============================================================================
# Verify installation
# =============================================================================

echo ""
info "Verifying installation..."

VERIFIED=0
for hook_file in "$HOOKS_DIR"/*; do
  if [ -f "$hook_file" ]; then
    HOOK_NAME=$(basename "$hook_file")
    if [ -x "$hook_file" ]; then
      success "Verified: $HOOK_NAME (executable)"
      VERIFIED=$((VERIFIED + 1))
    else
      error "Failed: $HOOK_NAME (not executable)"
    fi
  fi
done

# =============================================================================
# Summary
# =============================================================================

echo ""
echo -e "${BOLD}================================${NC}"

if [ "$VERIFIED" -eq "$HOOK_COUNT" ]; then
  success "All $HOOK_COUNT hook(s) installed successfully"
  echo ""
  info "Installed hooks:"
  for hook_file in "$HOOKS_DIR"/*; do
    if [ -f "$hook_file" ]; then
      echo "  - $(basename "$hook_file")"
    fi
  done
  echo ""
  info "Hooks will run automatically on git operations"
else
  error "Some hooks failed to install"
  exit 1
fi
