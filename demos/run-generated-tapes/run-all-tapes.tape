# Feature: Run generated tapes
# Scenario: Run all tapes

Source config.tape

Output demos/run-generated-tapes/run-all-tapes.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs run --all --output demos/"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
