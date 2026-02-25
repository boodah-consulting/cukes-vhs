# Feature: List available scenarios
# Scenario: List all translatable scenarios

Source config/config.tape

Output demos/list-available-scenarios/list-all-translatable-scenarios.gif

Hide

Show

Type@100ms "./cukes-vhs list --all"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
