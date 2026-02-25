# Feature: List available scenarios
# Scenario: Show scenario counts by source

Source config/config.tape

Output demos/output/list-available-scenarios/show-scenario-counts-by-source.gif
Output demos/output/list-available-scenarios/show-scenario-counts-by-source.ascii

Hide

Show

Type@100ms "./cukes-vhs list --count"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
