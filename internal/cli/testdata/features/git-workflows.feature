@git
Feature: Git Workflows
  As a developer
  I want to record git workflow demos
  So that new team members can learn our process

  Background:
    Given I am on the main menu

  @wip
  Scenario: Check repository status and log
    When I type "git status"
    And I press enter
    And I type "git log --oneline -5"
    And I press enter
    Then I should see the commit history

  Scenario: Stage changes interactively
    When I type "git add -p"
    And I press enter
    And I navigate down
    And I navigate down
    And I type "y"
    And I press enter

  Scenario: Search through commit log
    When I type "git log"
    And I press enter
    And I press "/" to search
    And I type "fix"
    And I press enter

  Scenario: Navigate log with vim keys
    When I press "j" to navigate down
    And I press "j" to navigate down
    And I press "k" to navigate up
    And I press enter
    And I press escape

  Scenario: Create a commit with message
    When I type "git commit -m 'fix: resolve login bug'"
    And I press enter
    Then I should see the commit confirmation

  Scenario: Interactive rebase workflow
    When I type "git rebase -i HEAD~3"
    And I press enter
    And I press "j" to navigate down
    And I type "squash"
    And I press escape
    Then the rebase should complete
