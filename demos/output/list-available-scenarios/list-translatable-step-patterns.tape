# Feature: List available scenarios
# Scenario: List translatable step patterns

Source config/config.tape

Output demos/output/list-available-scenarios/list-translatable-step-patterns.gif
Output demos/output/list-available-scenarios/list-translatable-step-patterns.ascii

Hide

Show

Type@100ms "./cukes-vhs list --steps"
Sleep 2s
Enter
Sleep 3s
Ctrl+C
