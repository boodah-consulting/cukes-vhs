# Feature: List available scenarios
# Scenario: Show scenario counts by source

Source config/config.tape

Output demos/list-available-scenarios/show-scenario-counts-by-source.gif

Hide

Show

Type@100ms "./cukes-vhs list --count"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
