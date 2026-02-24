Feature: List available scenarios
  As a developer using cukes-vhs
  I want to see which scenarios can be translated to VHS tapes
  So that I can choose what to generate

  Scenario: List all translatable scenarios
    When I type "./cukes-vhs list --all"
    And I press enter

  Scenario: List scenarios with JSON output
    When I type "./cukes-vhs list --all --json"
    And I press enter

  Scenario: List translatable step patterns
    When I type "./cukes-vhs list --steps"
    And I press enter

  Scenario: Show scenario counts by source
    When I type "./cukes-vhs list --count"
    And I press enter