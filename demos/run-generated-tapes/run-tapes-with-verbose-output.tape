# Feature: Run generated tapes
# Scenario: Run tapes with verbose output

Source config/config.tape

Output demos/run-generated-tapes/run-tapes-with-verbose-output.gif

Hide

Show

Type@100ms "./cukes-vhs run --all --output demos/ --verbose"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
