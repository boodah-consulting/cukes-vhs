# Feature: List available scenarios
# Scenario: List translatable step patterns

Source config.tape

Output demos/list-available-scenarios/list-translatable-step-patterns.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs list --steps"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
