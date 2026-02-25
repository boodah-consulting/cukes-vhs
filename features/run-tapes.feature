Feature: Run generated tapes
  As a developer using cukes-vhs
  I want to run the full tape pipeline
  So that I can generate, render, and validate tapes in one command

  Scenario: Run tape with cukes-vhs
    When I type "./cukes-vhs run --scenario 'list all translatable scenarios' --output demos/"
    And I press enter

  Scenario: Run all tapes
    When I type "./cukes-vhs run --all --output demos/"
    And I press enter

  Scenario: Run tapes with verbose output
    When I type "./cukes-vhs run --all --output demos/ --verbose"
    And I press enter
