Feature: Explore CLI help
  As a new user of cukes-vhs
  I want to explore the available commands and flags
  So that I understand how to use the tool

  Scenario: Display top-level help
    When I type "./cukes-vhs --help"
    And I press enter

  Scenario: Display list subcommand help
    When I type "./cukes-vhs list --help"
    And I press enter

  Scenario: Display generate subcommand help
    When I type "./cukes-vhs generate --help"
    And I press enter