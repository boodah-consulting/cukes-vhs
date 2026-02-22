#!/usr/bin/env bash
# AI-attributed commit script for cukes-vhs
# Non-interactive — suitable for CI and AI-assisted workflows
#
# Usage: bash scripts/ai-commit.sh <FILE> [NO_VERIFY] [AMEND]
#   FILE       — Path to file containing the commit message
#   NO_VERIFY  — Set to "1" to skip pre-commit hooks
#   AMEND      — Set to "1" to amend the previous commit
#
# Environment variables for agent detection:
#   Claude      — Set when running under Claude Code
#   OPENCODE    — Set when running under Opencode
#   CURSOR      — Set when running under Cursor

set -euo pipefail

# =============================================================================
# Colour helpers
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Colour

info()    { echo -e "${BLUE}ℹ${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn()    { echo -e "${YELLOW}⚠${NC} $*"; }
error()   { echo -e "${RED}✗${NC} $*" >&2; }

# =============================================================================
# Arguments
# =============================================================================

FILE="${1:-}"
NO_VERIFY="${2:-}"
AMEND="${3:-}"

if [ -z "$FILE" ]; then
  error "Usage: bash scripts/ai-commit.sh <FILE> [NO_VERIFY] [AMEND]"
  error "  FILE — path to a file containing the commit message"
  exit 1
fi

if [ ! -f "$FILE" ]; then
  error "Commit message file not found: $FILE"
  exit 1
fi

# =============================================================================
# Read commit message
# =============================================================================

COMMIT_MSG=$(cat "$FILE")

if [ -z "$COMMIT_MSG" ]; then
  error "Commit message file is empty: $FILE"
  exit 1
fi

info "Commit message read from: $FILE"

# =============================================================================
# Validate conventional commit format
# =============================================================================

HEADER=$(echo "$COMMIT_MSG" | head -1)

if ! echo "$HEADER" | grep -qE '^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\([a-z-]+\))?!?: .+'; then
  error "Commit message must follow conventional commit format"
  error "  Format: type(scope): description"
  error "  Valid types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
  error "  Got: $HEADER"
  exit 1
fi

success "Conventional commit format validated"

# =============================================================================
# Detect AI agent and model
# =============================================================================

format_model_name() {
  local model_id="$1"
  # Convert model identifiers to human-readable names
  # e.g. "claude-sonnet-4-6" → "Claude Sonnet 4 6"
  # e.g. "claude-opus-4-6" → "Claude Opus 4 6"
  # e.g. "gpt-5-mini" → "Gpt 5 Mini"
  echo "$model_id" \
    | sed 's/[-_]/ /g' \
    | sed 's/\b\(.\)/\u\1/g' \
    | sed 's/ \([0-9]\)/ \1/g'
}

detect_agent() {
  local agent_name=""
  local model_name=""

  # Claude Code detection
  if [ -n "${Claude:-}" ] || [ -n "${CLAUDE_CODE:-}" ]; then
    agent_name="Claude Code"
    # Try to detect model from environment
    if [ -n "${CLAUDE_MODEL:-}" ]; then
      model_name=$(format_model_name "$CLAUDE_MODEL")
    elif [ -n "${ANTHROPIC_MODEL:-}" ]; then
      model_name=$(format_model_name "$ANTHROPIC_MODEL")
    else
      model_name="Claude"
    fi
  # Opencode detection
  elif [ -n "${OPENCODE:-}" ]; then
    agent_name="Opencode"
    if [ -n "${OPENCODE_MODEL:-}" ]; then
      model_name=$(format_model_name "$OPENCODE_MODEL")
    else
      model_name="Claude"
    fi
  # Cursor detection
  elif [ -n "${CURSOR:-}" ] || [ -n "${CURSOR_SESSION:-}" ]; then
    agent_name="Cursor"
    if [ -n "${CURSOR_MODEL:-}" ]; then
      model_name=$(format_model_name "$CURSOR_MODEL")
    else
      model_name="Claude"
    fi
  # Generic AI detection
  elif [ -n "${AI_AGENT:-}" ]; then
    agent_name="${AI_AGENT}"
    if [ -n "${AI_MODEL:-}" ]; then
      model_name=$(format_model_name "$AI_MODEL")
    else
      model_name="Unknown Model"
    fi
  else
    error "No AI agent detected. Set one of: Claude, OPENCODE, CURSOR, or AI_AGENT"
    error "  Example: Claude=1 make ai-commit FILE=msg.txt"
    exit 1
  fi

  echo "${agent_name}|${model_name}"
}

AGENT_INFO=$(detect_agent)
AGENT_NAME=$(echo "$AGENT_INFO" | cut -d'|' -f1)
MODEL_NAME=$(echo "$AGENT_INFO" | cut -d'|' -f2)

info "Detected agent: ${CYAN}${AGENT_NAME}${NC} (${MODEL_NAME})"

# =============================================================================
# Check for staged changes (unless amending)
# =============================================================================

if [ "$AMEND" != "1" ]; then
  STAGED_COUNT=$(git diff --cached --name-only | wc -l | tr -d ' ')
  if [ "$STAGED_COUNT" -eq 0 ]; then
    error "No staged changes found. Stage files first with 'git add'"
    exit 1
  fi
  info "Staged files: $STAGED_COUNT"
fi

# =============================================================================
# Append AI attribution trailers
# =============================================================================

AI_TRAILER="AI-Generated-By: ${AGENT_NAME} (${MODEL_NAME})"
REVIEW_TRAILER="Reviewed-By: ${AGENT_NAME} (${MODEL_NAME})"

# Check if trailers already exist in the message
if echo "$COMMIT_MSG" | grep -q "AI-Generated-By:"; then
  info "AI-Generated-By trailer already present — skipping"
else
  # Ensure blank line before trailers
  if [ "$(echo "$COMMIT_MSG" | tail -1)" != "" ]; then
    COMMIT_MSG="${COMMIT_MSG}

${AI_TRAILER}"
  else
    COMMIT_MSG="${COMMIT_MSG}
${AI_TRAILER}"
  fi
  success "Added trailer: ${AI_TRAILER}"
fi

if echo "$COMMIT_MSG" | grep -q "Reviewed-By:"; then
  info "Reviewed-By trailer already present — skipping"
else
  COMMIT_MSG="${COMMIT_MSG}
${REVIEW_TRAILER}"
  success "Added trailer: ${REVIEW_TRAILER}"
fi

# =============================================================================
# Write final message to file
# =============================================================================

echo "$COMMIT_MSG" > "$FILE"

# =============================================================================
# Build git commit command
# =============================================================================

GIT_ARGS=("commit" "-F" "$FILE")

if [ "$NO_VERIFY" = "1" ]; then
  GIT_ARGS+=("--no-verify")
  warn "Skipping pre-commit hooks (NO_VERIFY=1)"
fi

if [ "$AMEND" = "1" ]; then
  GIT_ARGS+=("--amend")
  info "Amending previous commit (AMEND=1)"
fi

# =============================================================================
# Execute commit
# =============================================================================

info "Running: git ${GIT_ARGS[*]}"
echo ""

if git "${GIT_ARGS[@]}"; then
  echo ""
  success "Commit created successfully"
  echo ""
  info "Commit details:"
  git log -1 --format="  %h %s" HEAD
  echo ""
  # Show trailers
  TRAILERS=$(git log -1 --format="%(trailers)" HEAD 2>/dev/null || true)
  if [ -n "$TRAILERS" ]; then
    info "Trailers:"
    echo "$TRAILERS" | while IFS= read -r line; do
      [ -n "$line" ] && echo "  $line"
    done
  fi
else
  EXIT_CODE=$?
  error "Git commit failed (exit code: $EXIT_CODE)"
  exit "$EXIT_CODE"
fi
