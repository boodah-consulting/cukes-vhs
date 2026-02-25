# Feature: Generate VHS tape files
# Scenario: Generate tapes with verbose output

Source config/config.tape

Output demos/output/generate-vhs-tape-files/generate-tapes-with-verbose-output.gif
Output demos/output/generate-vhs-tape-files/generate-tapes-with-verbose-output.ascii

Hide

Show

Type@100ms "./cukes-vhs generate --all --output /tmp/tapes/ --verbose"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
