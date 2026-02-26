# Feature: Run generated tapes
# Scenario: Run tape with cukes-vhs

Source config/config.tape

Output demos/run-generated-tapes/run-tape-with-cukes-vhs.gif

Hide

Show

Type@100ms "./cukes-vhs run --scenario 'list all translatable scenarios' --output demos/"
Sleep 2s
Enter

Sleep 3s
Ctrl+C
