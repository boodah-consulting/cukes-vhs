# Feature: Run generated tapes
# Scenario: Run tapes with verbose output

Source config.tape

Output demos/run-generated-tapes/run-tapes-with-verbose-output.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs run --all --output demos/ --verbose"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
