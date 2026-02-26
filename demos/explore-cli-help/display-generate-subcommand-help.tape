# Feature: Explore CLI help
# Scenario: Display generate subcommand help

Source "config/config.tape"

Output "demos/explore-cli-help/display-generate-subcommand-help.gif"

Hide

Show

Type@100ms "./cukes-vhs generate --help"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
