Feature: TUI Application Navigation
  As a user of a terminal application
  I want to navigate menus and forms efficiently
  So that I can work productively

  Scenario: Navigate and search a list
    Given I am on the main menu
    When I press "/" to search
    And I type "config"
    And I press enter
    And I navigate down
    And I navigate down
    And I press enter

  Scenario: Vim-style navigation with input
    Given I am on the main menu
    When I press "j" to navigate down
    And I press "j" to navigate down
    And I press "k" to navigate up
    And I press enter
    And I press tab
    And I enter "new value"
    And I press Ctrl+S

  Scenario: Cancel and escape flows
    Given I am on the main menu
    When I navigate down
    And I press enter
    And I press escape
    And I navigate down
    And I press enter
    And I cancel
    And I close the modal
