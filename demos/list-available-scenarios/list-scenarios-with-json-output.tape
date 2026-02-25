# Feature: List available scenarios
# Scenario: List scenarios with JSON output

Source config/config.tape

Output demos/list-available-scenarios/list-scenarios-with-json-output.gif

Hide

Show

Type@100ms "./cukes-vhs list --all --json"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
