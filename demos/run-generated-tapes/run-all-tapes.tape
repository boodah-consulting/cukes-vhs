# Feature: Run generated tapes
# Scenario: Run all tapes

Source config/config.tape

Output demos/run-generated-tapes/run-all-tapes.gif

Hide

Show

Type@100ms "./cukes-vhs run --all --output demos/"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
