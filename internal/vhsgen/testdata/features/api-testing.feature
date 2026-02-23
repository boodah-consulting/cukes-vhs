@api
Feature: API Testing with curl
  As a backend developer
  I want to record API testing demos
  So that the team can test endpoints consistently

  Background:
    Given I am on the main menu

  Scenario: Health check endpoint
    When I type "curl -s http://localhost:8080/health | jq ."
    And I press enter
    Then I should see the health status

  Scenario: List all resources
    When I type "curl -s http://localhost:8080/api/v1/users | jq ."
    And I press enter
    And I navigate down
    And I press enter

  Scenario: Create a new resource
    When I enter "curl -X POST http://localhost:8080/api/v1/users -d '{\"name\": \"Alice\"}'"
    And I press enter
    Then the resource should be created

  Scenario: Search API responses
    When I type "curl -s http://localhost:8080/api/v1/users"
    And I press enter
    And I press "/" to search
    And I type "Alice"
    And I press enter

  Scenario: Navigate paginated results
    When I type "curl -s http://localhost:8080/api/v1/users?page=1"
    And I press enter
    And I press "j" to navigate down
    And I press "j" to navigate down
    And I press "k" to navigate up
    And I press escape

  Scenario Outline: Test different API endpoints
    When I type "curl -s http://localhost:8080/api/v1/<resource>"
    And I press enter
    Then I should see <count> results

    Examples:
      | resource | count |
      | users    | 10    |
      | orders   | 5     |
      | products | 20    |
