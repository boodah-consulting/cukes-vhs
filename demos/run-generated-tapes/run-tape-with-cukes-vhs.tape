# Feature: Run generated tapes
# Scenario: Run tape with cukes-vhs

Source config.tape

Output demos/run-generated-tapes/run-tape-with-cukes-vhs.gif

Hide
Type "# Verify cukes-vhs binary is available"
Enter
Type "command -v ./cukes-vhs || echo 'Binary not found'"
Enter
Sleep 1s
Show

Type@100ms "./cukes-vhs run --scenario 'list all translatable scenarios' --output demos/"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
