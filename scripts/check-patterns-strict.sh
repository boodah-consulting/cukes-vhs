#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VIOLATIONS=0

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 STRICT PATTERN ENFORCEMENT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

NEW_FILES=$(git diff --cached --name-only --diff-filter=A | grep '\.go$' | grep -v '_test.go' || true)
MODIFIED_FILES=$(git diff --cached --name-only --diff-filter=M | grep '\.go$' | grep -v '_test.go' || true)
ALL_STAGED=$(git diff --cached --name-only | grep '\.go$' || true)

echo "📁 New files: $(echo "$NEW_FILES" | grep -c '.' || echo 0)"
echo "📝 Modified files: $(echo "$MODIFIED_FILES" | grep -c '.' || echo 0)"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. DEPRECATED PATTERN CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

declare -A DEPRECATED_PATTERNS
DEPRECATED_PATTERNS=(
    ["components\.KeyBadge"]="primitives.HelpKeyBadge()"
    ["components\.StandardView"]="layout.NewScreenLayout()"
    ["components\.NewErrorModal"]="feedback.NewErrorModal()"
    ["components\.NewLoadingModal"]="feedback.NewLoadingModal()"
    ["components\.NewProgressModal"]="feedback.NewProgressModal()"
    ["components\.NewSuccessModal"]="feedback.NewSuccessModal()"
    ["components\.NewWarningModal"]="feedback.NewWarningModal()"
    ["components\.RenderOverlay"]="behaviors.RenderModalOverlay()"
    ["components\.RenderHelpFooter"]="primitives.RenderHelpFooter()"
    ["components\.ModalContainer"]="feedback.ModalContainer"
    ["components\.HelpModal"]="feedback.HelpModal"
    ["components\.ASCIILogo"]="DELETED - remove logo usage"
    ["CreateStandardView"]="layout.NewScreenLayout()"
    ["ThemedNavigationFooter"]="primitives.RenderHelpFooter()"
    ["ThemedCustomFooter"]="primitives.RenderHelpFooter()"
)

for file in $NEW_FILES; do
    for pattern in "${!DEPRECATED_PATTERNS[@]}"; do
        if grep -q "$pattern" "$file" 2>/dev/null; then
            echo -e "${RED}❌ NEW FILE uses deprecated pattern${NC}"
            echo "   File: $file"
            echo "   Pattern: $pattern"
            echo "   Use: ${DEPRECATED_PATTERNS[$pattern]}"
            VIOLATIONS=$((VIOLATIONS+1))
        fi
    done
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. HARDCODED COLORS CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $NEW_FILES; do
    HARDCODED=$(grep -n 'lipgloss\.Color("#' "$file" 2>/dev/null || true)
    if [ -n "$HARDCODED" ]; then
        echo -e "${RED}❌ NEW FILE has hardcoded colors${NC}"
        echo "   File: $file"
        echo "   Use theme.Primary(), theme.Secondary(), etc."
        echo "$HARDCODED" | head -3 | sed 's/^/   /'
        VIOLATIONS=$((VIOLATIONS+1))
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. FORBIDDEN COMMENT MARKERS CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $ALL_STAGED; do
    # Skip test files - they can reference bug numbers and issue markers
    if [[ ! "$file" =~ _test\.go$ ]]; then
        FORBIDDEN=$(grep -n 'TODO\|FIXME\|XXX\|HACK\|NOTE:\|BUG' "$file" 2>/dev/null || true)
        if [ -n "$FORBIDDEN" ]; then
            echo -e "${RED}❌ Forbidden comment markers found${NC}"
            echo "   File: $file"
            echo "   Forbidden: TODO, FIXME, XXX, HACK, NOTE, BUG"
            echo "   Use task tracking instead or refactor code"
            echo "$FORBIDDEN" | head -3 | sed 's/^/   /'
            VIOLATIONS=$((VIOLATIONS+1))
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. SKIPPED/PENDING TESTS CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

TEST_FILES=$(git diff --cached --name-only | grep '_test.go$' || true)

for file in $TEST_FILES; do
    SKIPPED=$(grep -n 'XIt(\|XDescribe(\|XContext(\|Skip(' "$file" 2>/dev/null || true)
    if [ -n "$SKIPPED" ]; then
        echo -e "${RED}❌ Skipped tests found${NC}"
        echo "   File: $file"
        echo "   Skipped tests are not allowed"
        echo "$SKIPPED" | head -3 | sed 's/^/   /'
        VIOLATIONS=$((VIOLATIONS+1))
    fi

    PENDING=$(grep -n 'PIt(\|PDescribe(\|PContext(' "$file" 2>/dev/null || true)
    if [ -n "$PENDING" ]; then
        echo -e "${RED}❌ Pending tests found${NC}"
        echo "   File: $file"
        echo "   Pending tests are not allowed"
        echo "$PENDING" | head -3 | sed 's/^/   /'
        VIOLATIONS=$((VIOLATIONS+1))
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. DEBUG STATEMENTS CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

DEBUG_PATTERNS='fmt\.Println\s*\(|fmt\.Printf\s*\(\s*"DEBUG|spew\.Dump|spew\.Printf|pp\.Print|pp\.Println|pretty\.Print'

for file in $ALL_STAGED; do
    if [[ ! "$file" =~ _test\.go$ ]]; then
        DEBUG=$(grep -nE "$DEBUG_PATTERNS" "$file" 2>/dev/null || true)
        if [ -n "$DEBUG" ]; then
            echo -e "${RED}❌ Debug statements found${NC}"
            echo "   File: $file"
            echo "   Remove debug statements before committing"
            echo "$DEBUG" | head -3 | sed 's/^/   /'
            VIOLATIONS=$((VIOLATIONS+1))
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. INLINE COMMENTS CHECK (NEW FILES ONLY)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $NEW_FILES; do
    # Skip doc.go files - they are entirely package documentation comments by design
    if [[ "$file" =~ doc\.go$ ]]; then
        continue
    fi

    # Check for end-of-line comments (code before //)
    EOL_COMMENTS=$(grep -nE '^[[:space:]]*[^/].*[[:space:]]+//' "$file" 2>/dev/null | grep -v '_test.go' | grep -v '//go:' || true)
    if [ -n "$EOL_COMMENTS" ]; then
        echo -e "${RED}❌ End-of-line comments found${NC}"
        echo "   File: $file"
        echo "   Comments must be on their own line above the code (except in test files)"
        echo "$EOL_COMMENTS" | head -3 | sed 's/^/   /'
        VIOLATIONS=$((VIOLATIONS+1))
    fi
    
    # Check for mid-function comments (comments not followed by type/func/var/const)
    MID_FUNC_COUNT=0
    while IFS= read -r line; do
        LINE_NUM=$(echo "$line" | cut -d: -f1)
        LINE_CONTENT=$(echo "$line" | cut -d: -f2-)
        
        # Skip godoc comments (// followed by capital letter or Package)
        if [[ "$LINE_CONTENT" =~ ^[[:space:]]*//[[:space:]][A-Z] ]] || [[ "$LINE_CONTENT" =~ ^[[:space:]]*//[[:space:]]Package ]]; then
            continue
        fi
        
        # Skip compiler directives
        if [[ "$LINE_CONTENT" =~ //go: ]] || [[ "$LINE_CONTENT" =~ //nolint ]]; then
            continue
        fi
        
        # Check next line - if it's not a declaration, it's mid-function
        NEXT_LINE_NUM=$((LINE_NUM + 1))
        NEXT_LINE=$(sed -n "${NEXT_LINE_NUM}p" "$file" 2>/dev/null || true)
        if [[ ! "$NEXT_LINE" =~ ^[[:space:]]*(func|type|var|const|package|import)[[:space:]] ]] && [[ ! "$NEXT_LINE" =~ ^[[:space:]]*$ ]]; then
            MID_FUNC_COUNT=$((MID_FUNC_COUNT+1))
        fi
    done < <(grep -n '^[[:space:]]*//[^/]' "$file" 2>/dev/null || true)

    if [ "$MID_FUNC_COUNT" -gt 0 ]; then
        echo -e "${RED}❌ Mid-function comments found${NC}"
        echo "   File: $file"
        echo "   Count: $MID_FUNC_COUNT"
        echo "   Extract to named methods instead of explaining with comments"
        VIOLATIONS=$((VIOLATIONS+1))
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "7. E2E TEST REQUIREMENTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

NEW_INTENTS=$(echo "$NEW_FILES" | grep '_intent\.go$' || true)
for intent in $NEW_INTENTS; do
    E2E_FILE="${intent%_intent.go}_e2e_test.go"
    if [ ! -f "$E2E_FILE" ]; then
        echo -e "${RED}❌ New intent missing E2E test${NC}"
        echo "   Intent: $intent"
        echo "   Required: $E2E_FILE"
        VIOLATIONS=$((VIOLATIONS+1))
    else
        if ! grep -q "Happy Paths" "$E2E_FILE"; then
            echo -e "${RED}❌ E2E missing Happy Paths section${NC}"
            echo "   File: $E2E_FILE"
            VIOLATIONS=$((VIOLATIONS+1))
        fi
        if ! grep -q "Sad Paths" "$E2E_FILE"; then
            echo -e "${RED}❌ E2E missing Sad Paths section${NC}"
            echo "   File: $E2E_FILE"
            VIOLATIONS=$((VIOLATIONS+1))
        fi
    fi
done

BUG_FILES=$(git diff --cached --name-only | grep 'bugs/BUG-' || true)
for bug in $BUG_FILES; do
    BUG_NUM=$(echo "$bug" | grep -oP 'BUG-\d+' || true)
    if [ -n "$BUG_NUM" ]; then
        # Find test files that mention the bug number in their diff
        REGRESSION_TEST=$(git diff --cached --name-only | grep '_test.go' | while read testfile; do
            git diff --cached "$testfile" | grep -q "$BUG_NUM" && echo "$testfile"
        done)
        if [ -z "$REGRESSION_TEST" ]; then
            echo -e "${RED}❌ Bug fix missing regression test${NC}"
            echo "   Bug: $BUG_NUM"
            echo "   Add E2E test with 'Bug Regressions' section mentioning $BUG_NUM"
            VIOLATIONS=$((VIOLATIONS+1))
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "8. DOCUMENTATION REQUIREMENTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -n "$NEW_INTENTS" ]; then
    WORKFLOW_DOCS=$(git diff --cached --name-only | grep 'docs/workflows/' || true)
    if [ -z "$WORKFLOW_DOCS" ]; then
        echo -e "${RED}❌ New intent requires workflow documentation${NC}"
        echo "   Update docs/workflows/ with new intent documentation"
        VIOLATIONS=$((VIOLATIONS+1))
    fi
fi

NEW_UIKIT=$(echo "$NEW_FILES" | grep 'internal/cli/uikit/' || true)
if [ -n "$NEW_UIKIT" ]; then
    UIKIT_DOC=$(git diff --cached --name-only | grep 'docs/UIKIT_GUIDE.md' || true)
    if [ -z "$UIKIT_DOC" ]; then
        echo -e "${RED}❌ New UIKit component requires UIKIT_GUIDE.md update${NC}"
        VIOLATIONS=$((VIOLATIONS+1))
    fi
fi

NEW_BEHAVIORS=$(echo "$NEW_FILES" | grep 'internal/cli/behaviors/' || true)
if [ -n "$NEW_BEHAVIORS" ]; then
    DEV_DOCS=$(git diff --cached --name-only | grep 'docs/development/' || true)
    if [ -z "$DEV_DOCS" ]; then
        echo -e "${RED}❌ New behavior requires development documentation${NC}"
        VIOLATIONS=$((VIOLATIONS+1))
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "9. BDD @wip TAG CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

FEATURE_FILES=$(git diff --cached --name-only | grep '\.feature$' || true)
if [ -n "$FEATURE_FILES" ]; then
    WIP_COUNT=0
    for file in $FEATURE_FILES; do
        FILE_WIP=$(grep -c '@wip' "$file" 2>/dev/null || echo 0)
        WIP_COUNT=$((WIP_COUNT + FILE_WIP))
    done
    
    if [ "$WIP_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  BDD scenarios with @wip tags: $WIP_COUNT${NC}"
        echo "   These scenarios are excluded from CI but should be completed"
        echo "   Run 'make bdd-wip' to see which scenarios need work"
        echo ""
        echo "   Files with @wip:"
        for file in $FEATURE_FILES; do
            COUNT=$(grep -c '@wip' "$file" 2>/dev/null || echo 0)
            if [ "$COUNT" -gt 0 ]; then
                echo "     $file: $COUNT scenario(s)"
            fi
        done
    else
        echo -e "${GREEN}✅ No @wip tags in staged feature files${NC}"
    fi
else
    echo -e "${GREEN}✅ No feature files in commit${NC}"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "10. COVERAGE CHECK"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -n "$ALL_STAGED" ]; then
    PACKAGES=$(echo "$ALL_STAGED" | xargs -I{} dirname {} | sort -u)
    for pkg in $PACKAGES; do
        # Skip mock packages - they are auto-generated and don't need tests
        if [[ "$pkg" == *"/mocks/"* ]] || [[ "$pkg" == *"/mocks" ]]; then
            echo -e "${YELLOW}⏭️  Skipping mock package: $pkg${NC}"
            continue
        fi
        # Skip BDD test infrastructure (features/ directory contains test code, not production code)
        if [[ "$pkg" == features* ]]; then
            echo -e "${GREEN}✅ $pkg: skipped (BDD test infrastructure)${NC}"
            continue
        fi
        # Skip test utilities (internal/testutil is test infrastructure, not production code)
        if [[ "$pkg" == *"/testutil"* ]] || [[ "$pkg" == *"/testutil/"* ]]; then
            echo -e "${GREEN}✅ $pkg: skipped (test infrastructure)${NC}"
            continue
        fi
        if [ -d "$pkg" ]; then
            COVERAGE_OUTPUT=$(go test -cover "./$pkg" 2>/dev/null || true)
            COVERAGE=$(echo "$COVERAGE_OUTPUT" | grep -oP 'coverage: \K[0-9.]+' || echo "0")
            if [ -n "$COVERAGE" ]; then
                # Round to nearest integer (94.5+ rounds to 95)
                COVERAGE_INT=$(printf "%.0f" "$COVERAGE" 2>/dev/null || echo "0")
                if [ "$COVERAGE_INT" -lt 95 ]; then
                    echo -e "${RED}❌ Package coverage below 95%${NC}"
                    echo "   Package: $pkg"
                    echo "   Coverage: $COVERAGE%"
                    VIOLATIONS=$((VIOLATIONS+1))
                else
                    echo -e "${GREEN}✅ $pkg: ${COVERAGE}%${NC}"
                fi
            fi
        fi
    done
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $VIOLATIONS -eq 0 ]; then
    echo -e "${GREEN}✅ All strict pattern checks passed${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${RED}❌ $VIOLATIONS violation(s) found${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
