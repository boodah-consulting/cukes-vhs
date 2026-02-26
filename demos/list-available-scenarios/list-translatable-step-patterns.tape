# Feature: List available scenarios
# Scenario: List translatable step patterns

Source config/config.tape

Output demos/list-available-scenarios/list-translatable-step-patterns.gif

Hide

Show

Type@100ms "./cukes-vhs list --steps"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
