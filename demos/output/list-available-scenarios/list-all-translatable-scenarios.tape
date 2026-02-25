# Feature: List available scenarios
# Scenario: List all translatable scenarios

Source config/config.tape

Output demos/output/list-available-scenarios/list-all-translatable-scenarios.gif
Output demos/output/list-available-scenarios/list-all-translatable-scenarios.ascii

Hide

Show

Type@100ms "./cukes-vhs list --all"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
