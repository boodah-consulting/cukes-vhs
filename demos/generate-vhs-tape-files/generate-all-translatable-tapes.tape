# Feature: Generate VHS tape files
# Scenario: Generate all translatable tapes

Source config.tape

Output demos/generate-vhs-tape-files/generate-all-translatable-tapes.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs generate --all --output /tmp/tapes/"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
