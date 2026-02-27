# Feature: Explore CLI help
# Scenario: Display list subcommand help

Source "config/config.tape"

Output "demos/explore-cli-help/display-list-subcommand-help.gif"

Hide

Show

Type@100ms "./cukes-vhs list --help"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
