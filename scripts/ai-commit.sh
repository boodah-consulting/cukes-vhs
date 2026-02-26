#!/bin/bash

set -euo pipefail

# ============================================================================
# AI Commit Helper
# ============================================================================
# Automates AI-attributed commits by:
# 1. Reading commit message from a file (or extracting from HEAD for amend)
# 2. Validating commit message format
# 3. Checking for staged changes (or unpushed HEAD for amend)
# 4. Adding AI attribution and human review trailers
# 5. Creating or amending the commit
#
# Usage: 
#   make ai-commit FILE=/path/to/commit-msg.txt
#   make ai-commit FILE=/path/to/commit-msg.txt NO_VERIFY=1
#   make ai-commit AMEND=1                       # Amend HEAD with AI attribution
#   make ai-commit AMEND=1 NO_VERIFY=1
#
# Environment variables for attribution override:
#   AI_AGENT - Override the AI agent name (auto-detected from OPENCODE env)
#   AI_MODEL - Override the model name (required if not auto-detected)
# ============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get file path from first argument
COMMIT_FILE="$1"

# Check if NO_VERIFY flag is set (passed as second argument)
NO_VERIFY="${2:-}"

# Check if AMEND flag is set (passed as third argument)
AMEND="${3:-}"

# ============================================================================
# Helper: Strip existing AI attribution from commit message
# ============================================================================
strip_ai_attribution() {
    local msg="$1"
    # Remove AI-Generated-By, Reviewed-By, and AI-Model lines (and blank lines before them)
    echo "$msg" | sed '/^$/N;/\nAI-Generated-By:/d' | \
        sed '/^AI-Generated-By:/d' | \
        sed '/^Reviewed-By:/d' | \
        sed '/^AI-Model:/d' | \
        sed -e :a -e '/^\n*$/{$d;N;ba' -e '}'  # Remove trailing blank lines
}

# ============================================================================
# AMEND MODE: Add AI attribution to existing HEAD commit
# ============================================================================
if [ "$AMEND" = "1" ]; then
    echo -e "${BLUE}🔄 AMEND MODE: Adding AI attribution to HEAD commit${NC}"
    echo ""
    
    # Safety check 1: Verify we have commits
    if ! git rev-parse HEAD &>/dev/null; then
        echo -e "${RED}❌ ERROR: No commits in repository${NC}"
        exit 1
    fi
    
    # Safety check 2: Verify HEAD hasn't been pushed
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    UPSTREAM=$(git rev-parse --abbrev-ref "@{upstream}" 2>/dev/null || echo "")
    
    if [ -n "$UPSTREAM" ]; then
        # Check if HEAD is ahead of upstream
        AHEAD=$(git rev-list --count "$UPSTREAM..HEAD" 2>/dev/null || echo "0")
        if [ "$AHEAD" = "0" ]; then
            echo -e "${RED}❌ ERROR: HEAD commit has already been pushed${NC}"
            echo ""
            echo "Cannot amend pushed commits. Options:"
            echo "  1. Create a new commit with the fix"
            echo "  2. Use 'git push --force' (DANGEROUS - only if you're certain)"
            echo ""
            exit 1
        fi
        echo -e "${GREEN}✅ HEAD is unpushed (${AHEAD} commit(s) ahead of ${UPSTREAM})${NC}"
    else
        echo -e "${YELLOW}⚠️  No upstream branch - assuming commit is unpushed${NC}"
    fi
    
    # Safety check 3: Check if commit already has AI attribution
    CURRENT_MSG=$(git log -1 --pretty=%B)
    if echo "$CURRENT_MSG" | grep -q "^AI-Generated-By:"; then
        echo -e "${YELLOW}⚠️  HEAD commit already has AI attribution${NC}"
        echo ""
        echo "Current attribution:"
        echo "$CURRENT_MSG" | grep -E "^(AI-Generated-By|Reviewed-By|AI-Model):" | sed 's/^/  /'
        echo ""
        read -p "Replace existing attribution? [y/N] " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Aborted."
            exit 0
        fi
    fi
    
    # Extract original message (strip any existing attribution)
    COMMIT_MSG=$(strip_ai_attribution "$CURRENT_MSG")
    
    echo -e "${BLUE}📄 Extracted commit message from HEAD:${NC}"
    echo "─────────────────────────────────────────────"
    echo "$COMMIT_MSG"
    echo "─────────────────────────────────────────────"
    
    # Get first line for validation
    FIRST_LINE=$(echo "$COMMIT_MSG" | head -n1 | sed 's/[[:space:]]*$//')
    
    # Skip to Step 3 (validation) - no staged changes check needed for amend
    
else
    # ============================================================================
    # NORMAL MODE: Create new commit
    # ============================================================================
    
    # Step 1: Validate file provided and read commit message
    if [ -z "$COMMIT_FILE" ]; then
        echo -e "${RED}❌ ERROR: Commit message file required${NC}"
        echo ""
        echo "Usage:"
        echo "  make ai-commit FILE=/path/to/commit-msg.txt"
        echo "  make ai-commit AMEND=1                        # Amend HEAD with AI attribution"
        echo ""
        echo "Create your commit message file first:"
        echo ""
        echo "  cat > tmp/commit.txt << 'EOF'"
        echo "  feat(scope): short description"
        echo ""
        echo "  Optional longer explanation..."
        echo "  EOF"
        echo ""
        echo "  make ai-commit FILE=tmp/commit.txt"
        echo ""
        exit 1
    fi

    # Check if file exists and is readable
    if [ ! -f "$COMMIT_FILE" ]; then
        echo -e "${RED}❌ ERROR: File not found: ${COMMIT_FILE}${NC}"
        echo ""
        echo "Create the file first:"
        echo "  cat > ${COMMIT_FILE} << 'EOF'"
        echo "  feat(scope): description"
        echo "  EOF"
        echo ""
        exit 1
    fi

    if [ ! -r "$COMMIT_FILE" ]; then
        echo -e "${RED}❌ ERROR: Cannot read file: ${COMMIT_FILE}${NC}"
        exit 1
    fi

    echo -e "${BLUE}📄 Reading commit message from: ${COMMIT_FILE}${NC}"
    COMMIT_MSG=$(cat "$COMMIT_FILE")

    # Validate we have a message
    if [ -z "$COMMIT_MSG" ]; then
        echo -e "${RED}❌ ERROR: Commit message file is empty${NC}"
        exit 1
    fi

    # Validate message is not a placeholder
    if [[ "$COMMIT_MSG" =~ ^\.\.\.$ ]] || [[ "$COMMIT_MSG" =~ ^\.\.\.\s*$ ]] || [[ "$COMMIT_MSG" == "..." ]]; then
        echo -e "${RED}❌ ERROR: Commit message cannot be '...' placeholder${NC}"
        echo ""
        echo "Edit your file with an actual commit message:"
        echo "  ${COMMIT_FILE}"
        echo ""
        exit 1
    fi

    # Validate message has actual content (not just type prefix)
    # Get first line and strip whitespace
    FIRST_LINE=$(echo "$COMMIT_MSG" | head -n1 | sed 's/[[:space:]]*$//')
    if [[ "$FIRST_LINE" =~ ^[a-z]+\([a-zA-Z0-9_-]+\):$ ]] || [[ "$FIRST_LINE" =~ ^[a-z]+:$ ]]; then
        echo -e "${RED}❌ ERROR: Commit message has no description${NC}"
        echo ""
        echo "Edit your file to add a description after the colon:"
        echo "  ${COMMIT_FILE}"
        echo ""
        exit 1
    fi

    # Step 2: Check for staged changes
    echo ""
    echo -e "${BLUE}🔍 Checking for staged changes...${NC}"

    if git diff --cached --quiet; then
        echo -e "${RED}❌ ERROR: No staged changes${NC}"
        echo ""
        echo "You must stage changes before committing:"
        echo "  git add -p <file>          # Stage specific hunks interactively"
        echo "  git add <file>             # Stage entire file"
        echo ""
        echo "Then try again:"
        echo "  make ai-commit FILE=${COMMIT_FILE}"
        echo ""
        exit 1
    fi

    echo -e "${GREEN}✅ Staged changes detected${NC}"
fi

# ============================================================================
# Step 3: Validate commit message format
# ============================================================================

echo ""
echo -e "${BLUE}🔍 Validating commit message format...${NC}"

# Basic conventional commit format check
if ! echo "$FIRST_LINE" | grep -qE "^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\([a-zA-Z0-9_-]+\))?: .+"; then
    echo -e "${YELLOW}⚠️  Warning: Message may not follow conventional commit format${NC}"
    echo ""
    echo "Recommended format:"
    echo -e "${GREEN}  type(scope): subject${NC}"
    echo ""
    echo "Examples:"
    echo "  feat(chat): add streaming response"
    echo "  fix(nav): resolve scroll issue"
    echo "  docs(readme): update installation"
    echo ""
else
    echo -e "${GREEN}✅ Commit message format valid${NC}"
fi

# ============================================================================
# Step 4: Detect AI agent and model
# ============================================================================

# Auto-detect AI agent from environment
detect_ai_agent() {
    echo "Opencode"
}

# Detect model - REQUIRED, no defaults
detect_ai_model() {
    # Check for explicit override first
    if [ -n "$AI_MODEL" ]; then
        echo "$AI_MODEL"
        return
    fi
    
    # No model detected
    echo ""
}

# Format model name to human-readable format
# Converts hyphenated model IDs to Title Case with proper spacing
# Examples: claude-sonnet-4-5 -> Claude Sonnet 4.5
#           gpt-4o -> GPT-4o
#           llama3-70b -> Llama3 70B
format_model_name() {
    local model="$1"
    
    # Special case for gpt-4o format (preserve -4o as hyphenated)
    if [[ "$model" =~ ^gpt-[0-9]+o$ ]]; then
        echo "$model" | sed 's/gpt/GPT/'
        return
    fi
    
    # Replace hyphens with spaces
    local formatted="${model//-/ }"
    
    # Capitalize first letter of each word
    formatted=$(echo "$formatted" | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')
    
    # Fix version numbers: "4 5" -> "4.5", "5 1" -> "5.1"
    # Only for standalone single-digit numbers (use word boundaries)
    formatted=$(echo "$formatted" | sed -E 's/\b([0-9]) ([0-9])\b/\1.\2/g')
    
    # Uppercase size suffixes (70b -> 70B, 8b -> 8B, 7b -> 7B)
    formatted=$(echo "$formatted" | sed -E 's/([0-9]+)b$/\1B/g')
    
    # Uppercase special prefixes
    formatted=$(echo "$formatted" | sed 's/^Gpt/GPT/g')
    
    echo "$formatted"
}

AGENT_NAME=$(detect_ai_agent)
MODEL_NAME=$(format_model_name "$(detect_ai_model)")

# Validate agent detected
if [ -z "$AGENT_NAME" ]; then
    echo -e "${RED}❌ ERROR: Could not detect AI agent${NC}"
    echo ""
    echo "Set the AI_AGENT environment variable:"
    echo "  export AI_AGENT='Opencode'"
    echo ""
    echo "Or run with:"
    echo "  AI_AGENT='Opencode' AI_MODEL='claude-opus-4-5' make ai-commit FILE=tmp/commit.txt"
    echo ""
    exit 1
fi

# Validate model - REQUIRED
if [ -z "$MODEL_NAME" ]; then
    echo -e "${RED}❌ ERROR: AI_MODEL environment variable not set${NC}"
    echo ""
    echo "The model must be specified for accurate attribution."
    echo ""
    echo "Set the AI_MODEL environment variable:"
    echo "  export AI_MODEL='claude-opus-4-5'"
    echo ""
    echo "Or run with:"
    echo "  AI_MODEL='claude-opus-4-5' make ai-commit FILE=tmp/commit.txt"
    echo ""
    echo "Common models:"
    echo "  claude-opus-4-5, claude-sonnet-4, gpt-4o, llama3.2"
    echo ""
    exit 1
fi

# Get reviewer name and email from git config
REVIEWER_NAME=$(git config user.name)
REVIEWER_EMAIL=$(git config user.email)

if [ -z "$REVIEWER_NAME" ] || [ -z "$REVIEWER_EMAIL" ]; then
    echo -e "${YELLOW}⚠️  Warning: git user.name or user.email not set${NC}"
    echo "Set them with: git config user.name \"Your Name\" && git config user.email \"your.email@example.com\""
    REVIEWER_NAME="${REVIEWER_NAME:-Unknown}"
    REVIEWER_EMAIL="${REVIEWER_EMAIL:-unknown@example.com}"
fi

# ============================================================================
# Step 5: Create or amend commit with AI attribution
# ============================================================================

echo ""
if [ "$AMEND" = "1" ]; then
    echo -e "${BLUE}🤖 Amending commit with AI attribution...${NC}"
else
    echo -e "${BLUE}🤖 Creating AI-attributed commit...${NC}"
fi
echo ""
echo "Agent:    ${AGENT_NAME}"
echo "Model:    ${MODEL_NAME}"
echo "Reviewer: ${REVIEWER_NAME}"
echo ""

# Build full commit message with attribution
FINAL_MSG_FILE=$(mktemp)

cat > "$FINAL_MSG_FILE" <<EOF
${COMMIT_MSG}

AI-Generated-By: ${AGENT_NAME} (${MODEL_NAME})
Reviewed-By: ${REVIEWER_NAME} <${REVIEWER_EMAIL}>
EOF

# Build commit flags array
COMMIT_FLAGS=()
COMMIT_FLAGS+=("-F" "$FINAL_MSG_FILE")

# Add --amend flag if in AMEND mode
if [ "$AMEND" = "1" ]; then
    COMMIT_FLAGS+=("--amend")
fi

# Add --no-verify flag if NO_VERIFY is set
if [ "$NO_VERIFY" = "1" ]; then
    echo -e "${YELLOW}⚠️  Skipping pre-commit hooks (--no-verify)${NC}"
    echo ""
    COMMIT_FLAGS+=("--no-verify")
fi

if git commit "${COMMIT_FLAGS[@]}"; then
    echo ""
    if [ "$AMEND" = "1" ]; then
        echo -e "${GREEN}✅ Commit amended successfully${NC}"
    else
        echo -e "${GREEN}✅ Commit created successfully${NC}"
    fi
    echo ""
    echo "Commit message:"
    echo "─────────────────────────────────────────────"
    git log -1 --pretty=%B
    echo "─────────────────────────────────────────────"
    echo ""
    
    rm -f "$FINAL_MSG_FILE"
else
    echo ""
    echo -e "${RED}❌ Commit failed${NC}"
    rm -f "$FINAL_MSG_FILE"
    exit 1
fi

# ============================================================================
# Step 6: Summary
# ============================================================================

if [ "$AMEND" = "1" ]; then
    echo -e "${GREEN}✅ AI attribution added to commit${NC}"
else
    echo -e "${GREEN}✅ AI-attributed commit complete${NC}"
fi
echo ""
echo "Next steps:"
echo "  git log -1                  # Review the commit"
echo "  make check-compliance       # Run compliance checks"
if [ "$AMEND" = "1" ]; then
    echo "  git push --force-with-lease # Push amended commit (if already pushed)"
else
    echo "  git push                    # Push to remote (when ready)"
fi
echo ""
