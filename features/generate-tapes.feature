Feature: Generate VHS tape files
  As a developer using cukes-vhs
  I want to generate VHS tape files from my Cucumber scenarios
  So that I can create terminal recordings of my application

  Scenario: Generate all translatable tapes
    When I type "./cukes-vhs generate --all --output /tmp/tapes/"
    And I press enter

  Scenario: Generate tapes for a specific feature
    When I type "./cukes-vhs generate --feature 'List available scenarios' --output /tmp/tapes/"
    And I press enter

  Scenario: Generate tapes with verbose output
    When I type "./cukes-vhs generate --all --output /tmp/tapes/ --verbose"
    And I press enter