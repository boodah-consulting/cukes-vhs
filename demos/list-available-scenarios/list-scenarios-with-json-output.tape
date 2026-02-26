# Feature: List available scenarios
# Scenario: List scenarios with JSON output

Source config.tape

Output demos/list-available-scenarios/list-scenarios-with-json-output.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs list --all --json"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
