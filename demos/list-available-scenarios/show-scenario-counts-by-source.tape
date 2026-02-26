# Feature: List available scenarios
# Scenario: Show scenario counts by source

Source config.tape

Output demos/list-available-scenarios/show-scenario-counts-by-source.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs list --count"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
