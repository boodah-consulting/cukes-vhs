# Feature: Generate VHS tape files
# Scenario: Generate tapes for a specific feature

Source config/config.tape

Output demos/output/generate-vhs-tape-files/generate-tapes-for-a-specific-feature.gif
Output demos/output/generate-vhs-tape-files/generate-tapes-for-a-specific-feature.ascii

Hide

Show

Type@100ms "./cukes-vhs generate --feature 'List available scenarios' --output /tmp/tapes/"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
