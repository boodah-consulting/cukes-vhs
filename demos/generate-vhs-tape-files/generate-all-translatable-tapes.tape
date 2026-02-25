# Feature: Generate VHS tape files
# Scenario: Generate all translatable tapes

Source config/config.tape

Output demos/generate-vhs-tape-files/generate-all-translatable-tapes.gif

Hide

Show

Type@100ms "./cukes-vhs generate --all --output /tmp/tapes/"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
