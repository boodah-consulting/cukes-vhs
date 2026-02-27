# Feature: Generate VHS tape files
# Scenario: Generate tapes with verbose output

Source "config/config.tape"

Output "demos/generate-vhs-tape-files/generate-tapes-with-verbose-output.gif"

Hide

Show

Type@100ms "./cukes-vhs generate --all --output /tmp/tapes/ --verbose"
Sleep 2s
Enter
Sleep 2s

Sleep 3s
Ctrl+C
