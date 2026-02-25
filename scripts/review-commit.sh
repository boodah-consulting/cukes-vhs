#!/bin/bash

set -e

echo "================================================"
echo "🔍 ATOMIC COMMIT REVIEW"
echo "================================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if there are staged changes
if ! git diff --cached --quiet; then
    echo -e "${GREEN}✅ Staged changes detected${NC}"
else
    echo -e "${RED}❌ No staged changes${NC}"
    echo ""
    echo "Stage your changes first with: git add <files>"
    exit 1
fi

echo ""
echo "------------------------------------------------"
echo "📋 STAGED FILES"
echo "------------------------------------------------"
git diff --cached --name-status

echo ""
echo "------------------------------------------------"
echo "📊 CHANGE STATISTICS"
echo "------------------------------------------------"
git diff --cached --stat

# Count files and lines
FILE_COUNT=$(git diff --cached --name-only | wc -l | tr -d ' ')
LINE_COUNT=$(git diff --cached --numstat | awk '{add+=$1; del+=$2} END {print add+del}')

echo ""
echo -e "${BLUE}Files changed: $FILE_COUNT${NC}"
echo -e "${BLUE}Lines changed: $LINE_COUNT${NC}"

# Warnings
if [ "$FILE_COUNT" -gt 10 ]; then
    echo -e "${YELLOW}⚠️  Warning: More than 10 files changed. Consider splitting.${NC}"
fi

if [ "$LINE_COUNT" -gt 500 ]; then
    echo -e "${YELLOW}⚠️  Warning: More than 500 lines changed. Consider splitting (unless initial setup).${NC}"
fi

# Check for generated files
echo ""
echo "------------------------------------------------"
echo "🔍 CHECKING FOR GENERATED FILES"
echo "------------------------------------------------"
GENERATED_FILES=$(git diff --cached --name-only | grep -E '\.(out|exe|dll|so|dylib|test)$' || true)
if [ -n "$GENERATED_FILES" ]; then
    echo -e "${RED}❌ Generated files detected! Remove from staging:${NC}"
    echo "$GENERATED_FILES"
    echo ""
    echo "Run: git reset <filename>"
    exit 1
else
    echo -e "${GREEN}✅ No generated files detected${NC}"
fi

# Check for coverage files
COVERAGE_FILES=$(git diff --cached --name-only | grep -E 'coverage\.(out|html)' || true)
if [ -n "$COVERAGE_FILES" ]; then
    echo -e "${RED}❌ Coverage files detected! Remove from staging:${NC}"
    echo "$COVERAGE_FILES"
    echo ""
    echo "Run: git reset coverage.out coverage.html"
    exit 1
fi

# Check for debug statements
echo ""
echo "------------------------------------------------"
echo "🐛 CHECKING FOR DEBUG CODE"
echo "------------------------------------------------"
DEBUG_PATTERNS='TODO|FIXME|XXX|console\.log|debugger|fmt\.Println\("DEBUG'
if git diff --cached | grep -E "$DEBUG_PATTERNS" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Warning: Debug statements found:${NC}"
    git diff --cached | grep -E "$DEBUG_PATTERNS" | head -5
    echo ""
    echo "Consider removing debug code before committing."
else
    echo -e "${GREEN}✅ No debug statements found${NC}"
fi

# Check for secrets
echo ""
echo "------------------------------------------------"
echo "🔐 CHECKING FOR SECRETS"
echo "------------------------------------------------"
SECRET_PATTERNS='password["\s]*[:=]|secret["\s]*[:=]|api[_-]?key["\s]*[:=]|token["\s]*[:=]|credential["\s]*[:=]'
if git diff --cached | grep -iE "$SECRET_PATTERNS" | grep -v 'password_hash' > /dev/null 2>&1; then
    echo -e "${RED}❌ Potential secrets detected!${NC}"
    git diff --cached | grep -iE "$SECRET_PATTERNS" | head -5
    echo ""
    echo "DO NOT commit secrets! Use environment variables instead."
    exit 1
else
    echo -e "${GREEN}✅ No secrets detected${NC}"
fi

# Build check
echo ""
echo "------------------------------------------------"
echo "🏗️  BUILD CHECK"
echo "------------------------------------------------"
if go build ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Build successful${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    echo ""
    echo "Fix build errors before committing:"
    go build ./...
    exit 1
fi

# Test check
echo ""
echo "------------------------------------------------"
echo "🧪 TEST CHECK"
echo "------------------------------------------------"
if go test ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✅ All tests pass${NC}"
else
    echo -e "${RED}❌ Tests failed${NC}"
    echo ""
    echo "Fix failing tests before committing:"
    go test ./...
    exit 1
fi

# Race check
echo ""
echo "------------------------------------------------"
echo "🏁 RACE CONDITION CHECK"
echo "------------------------------------------------"
if go test -race ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✅ No race conditions detected${NC}"
else
    echo -e "${RED}❌ Race conditions detected${NC}"
    echo ""
    echo "Fix race conditions before committing:"
    go test -race ./...
    exit 1
fi

# Formatting check
echo ""
echo "------------------------------------------------"
echo "✨ FORMATTING CHECK"
echo "------------------------------------------------"
UNFORMATTED=$(gofmt -l . 2>&1 | grep -v '^vendor/' | grep '\.go$' || true)
if [ -z "$UNFORMATTED" ]; then
    echo -e "${GREEN}✅ All files properly formatted${NC}"
else
    echo -e "${YELLOW}⚠️  Unformatted files:${NC}"
    echo "$UNFORMATTED"
    echo ""
    echo "Run: go fmt ./..."
fi

# Check architectural layers
echo ""
echo "------------------------------------------------"
echo "🏛️  ARCHITECTURAL LAYERS AFFECTED"
echo "------------------------------------------------"
LAYERS=$(git diff --cached --name-only | grep -E 'internal/(domain|service|repository|cli|logger)' | cut -d'/' -f2-3 | sort -u || true)
if [ -n "$LAYERS" ]; then
    echo "$LAYERS" | while read layer; do
        echo -e "${BLUE}  - $layer${NC}"
    done

    LAYER_COUNT=$(echo "$LAYERS" | wc -l | tr -d ' ')
    if [ "$LAYER_COUNT" -gt 2 ]; then
        echo ""
        echo -e "${YELLOW}⚠️  Multiple layers affected ($LAYER_COUNT). Consider splitting by layer.${NC}"
    fi
else
    echo -e "${BLUE}Non-code changes (docs, config, etc.)${NC}"
fi

# Check for AI attribution
echo ""
echo "------------------------------------------------"
echo "🤖 AI ATTRIBUTION CHECK"
echo "------------------------------------------------"

CODE_FILES_CHECK=$(git diff --cached --name-only | grep -E '\.(go|js|ts|py|java|c|cpp|rs)$' || true)

if [ -n "$CODE_FILES_CHECK" ]; then
    echo -e "${YELLOW}⚠️  Code files detected in commit${NC}"
    echo ""
    echo "If ANY code was AI-generated, your commit message MUST include:"
    echo ""
    echo -e "${BLUE}  AI-Generated-By: <Assistant Name> (<Model Version>)${NC}"
    echo -e "${BLUE}  Reviewed-By: <Your Name>${NC}"
    echo ""
    echo "Examples:"
    echo "  AI-Generated-By: Avante (Claude 3.5 Sonnet)"
    echo "  AI-Generated-By: Claude (Claude 3.7 Sonnet)"
    echo "  AI-Generated-By: GitHub Copilot (GPT-4)"
    echo ""
    echo -e "${YELLOW}See docs/rules/AI_COMMIT_ATTRIBUTION.md for full guidelines${NC}"
else
    echo -e "${GREEN}✅ No code files in this commit${NC}"
fi

# Final checklist
echo ""
echo "================================================"
echo "✅ COMMIT READINESS CHECKLIST"
echo "================================================"
echo ""
echo "Before committing, verify:"
echo ""
echo "  [ ] This commit represents ONE logical change"
echo "  [ ] Commit message follows conventional format"
echo "  [ ] Commit message explains WHY, not just WHAT"
echo "  [ ] All tests pass (verified above ✓)"
echo "  [ ] Code is properly formatted (verified above ✓)"
echo "  [ ] No generated files included (verified above ✓)"
echo "  [ ] No secrets or sensitive data included (verified above ✓)"
echo "  [ ] Related tests are included"
echo "  [ ] Documentation updated (if needed)"
echo "  [ ] AI attribution included (if AI-generated code) 🤖"
echo ""
echo "------------------------------------------------"
echo "📝 COMMIT MESSAGE FORMAT:"
echo "------------------------------------------------"
echo ""
echo -e "${BLUE}<type>(<scope>): <subject>${NC}"
echo ""
echo "<body explaining why this change is needed>"
echo ""
echo "<footer with issue references>"
echo ""
echo "Types: feat, fix, docs, style, refactor, test, chore, perf"
echo "Scopes: domain, service, repo, cli, logger"
echo ""
echo -e "${GREEN}Example:${NC}"
echo "feat(service): add event filtering by date range"
echo ""
echo "Implement date range filtering for timeline views."
echo "Improves performance for large event collections."
echo ""
echo "Closes #56"
echo ""
echo "================================================"
echo ""

# Summary
echo -e "${GREEN}✅ All automated checks passed!${NC}"
echo ""
echo "Review the checklist above and proceed with your commit."
echo ""
echo "To commit: git commit"
echo "To review diff: git diff --cached"
echo "To unstage all: git reset"
echo ""

