#!/bin/bash

set -e

TASK_NAME="$1"

if [ -z "$TASK_NAME" ]; then
    echo "Usage: $0 'feature name'"
    exit 1
fi

TASKS_DIR="tasks"
TEMPLATE="docs/templates/task-template.md"

mkdir -p "$TASKS_DIR"

LAST_NUM=$(ls -1 "$TASKS_DIR"/TASK-*.md 2>/dev/null | grep -oP 'TASK-\K\d+' | sort -n | tail -1 || echo "0")
NEXT_NUM=$((LAST_NUM + 1))
TASK_ID=$(printf "TASK-%03d" $NEXT_NUM)

SLUG=$(echo "$TASK_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd 'a-z0-9-')
FILENAME="${TASKS_DIR}/${TASK_ID}-${SLUG}.md"

if [ -f "$TEMPLATE" ]; then
    sed "s/TASK-XXX/${TASK_ID}/g; s/\[Title\]/${TASK_NAME}/g" "$TEMPLATE" > "$FILENAME"
else
    cat > "$FILENAME" << EOF
# ${TASK_ID}: ${TASK_NAME}

## Summary

Brief description of the task.

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2

## Technical Notes

### Files to Modify

- \`path/to/file.go\` - Description

### Patterns to Use

Run \`make what-to-use NEED="keyword"\` for component guidance.

## Testing Requirements

- [ ] Unit tests
- [ ] E2E tests (if new intent)

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Tests written FIRST (TDD)
- [ ] Tests pass with >= 95% coverage
- [ ] No pattern violations
- [ ] Compliance check passes
- [ ] Committed with \`make ai-commit\`
EOF
fi

echo "Created: $FILENAME"
echo ""
echo "Next steps:"
echo "  1. Edit the task file with details"
echo "  2. Run: make pre-task"
echo "  3. Start TDD: make tdd-red"
