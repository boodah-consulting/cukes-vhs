#!/bin/bash

set -e

BUG_DESC="$1"

if [ -z "$BUG_DESC" ]; then
    echo "Usage: $0 'bug description'"
    exit 1
fi

BUGS_DIR="bugs"
TEMPLATE="docs/templates/bug-template.md"

mkdir -p "$BUGS_DIR"

LAST_NUM=$(ls -1 "$BUGS_DIR"/BUG-*.md 2>/dev/null | grep -oP 'BUG-\K\d+' | sort -n | tail -1 || echo "0")
NEXT_NUM=$((LAST_NUM + 1))
BUG_ID=$(printf "BUG-%03d" $NEXT_NUM)

SLUG=$(echo "$BUG_DESC" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd 'a-z0-9-')
FILENAME="${BUGS_DIR}/${BUG_ID}-${SLUG}.md"

if [ -f "$TEMPLATE" ]; then
    sed "s/BUG-XXX/${BUG_ID}/g; s/\[Title\]/${BUG_DESC}/g" "$TEMPLATE" > "$FILENAME"
else
    cat > "$FILENAME" << EOF
# ${BUG_ID}: ${BUG_DESC}

## Summary

Brief description of the bug.

## Steps to Reproduce

1. Step 1
2. Step 2

## Expected Behavior

What should happen.

## Actual Behavior

What actually happens.

## Environment

- OS: 
- Go version: 
- Branch: 

## Severity

- [ ] Critical
- [ ] High
- [ ] Medium
- [ ] Low

## Notes

Additional context.
EOF
fi

echo "Created: $FILENAME"
echo ""
echo "Next steps:"
echo "  1. Fill in bug details"
echo "  2. Create fix task: make new-feature TASK='Fix ${BUG_ID}'"
echo "  3. Write regression test FIRST"
