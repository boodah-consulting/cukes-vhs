# Feature: Explore CLI help
# Scenario: Display generate subcommand help

Source config.tape

Output demos/explore-cli-help/display-generate-subcommand-help.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs generate --help"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
