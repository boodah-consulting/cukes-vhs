# Feature: Explore CLI help
# Scenario: Display top-level help

Source config/config.tape

Output demos/explore-cli-help/display-top-level-help.gif

Hide

Show

Type@100ms "./cukes-vhs --help"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
