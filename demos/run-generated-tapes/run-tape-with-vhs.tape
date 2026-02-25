# Feature: Run generated tapes
# Scenario: Run tape with vhs

Source config/config.tape

Output demos/run-generated-tapes/run-tape-with-vhs.gif

Hide

Show

Type@100ms "vhs demos/output/list-available-scenarios/list-all-translatable-scenarios.tape"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
