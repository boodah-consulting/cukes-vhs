#!/usr/bin/env python3
"""Fix AI attribution trailers in git commit messages."""
import sys
import re

lines = sys.stdin.read().splitlines()

output_lines = []
seen_ai_generated = False
seen_reviewed_by = False

for line in lines:
    if line.startswith('AI-Generated-By:'):
        if seen_ai_generated:
            continue
        seen_ai_generated = True
        m = re.search(r'\(([^)]+)\)\s*$', line)
        if m:
            model = m.group(1)
            line = f'AI-Generated-By: Opencode ({model})'
        output_lines.append(line)

    elif line.startswith('Reviewed-By:'):
        if seen_reviewed_by:
            continue
        seen_reviewed_by = True
        output_lines.append('Reviewed-By: Yomi Colledge')

    elif line.startswith('AI-Model:'):
        continue

    else:
        output_lines.append(line)

while output_lines and output_lines[-1].strip() == '':
    output_lines.pop()

if not seen_ai_generated:
    if output_lines and output_lines[-1].strip() != '':
        output_lines.append('')
    output_lines.append('AI-Generated-By: Opencode (Claude Opus 4.5)')
    output_lines.append('Reviewed-By: Yomi Colledge')
elif not seen_reviewed_by:
    output_lines.append('Reviewed-By: Yomi Colledge')

print('\n'.join(output_lines))
