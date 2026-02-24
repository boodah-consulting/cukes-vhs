Feature: Docker Management
  As a DevOps engineer
  I want to record container management demos
  So that the team can follow standard procedures

  Scenario: List running containers
    Given the database is empty
    When I type "docker ps"
    And I press enter
    And I type "docker images"
    And I press enter

  Scenario: Inspect a container
    Given the database is empty
    And I am on the main menu
    When I type "docker inspect myapp"
    And I press enter
    And I navigate down
    And I navigate down
    And I press enter

  Scenario: View container logs
    Given I am on the main menu
    When I type "docker logs -f myapp"
    And I press enter
    And I press escape

  Scenario: Build and tag an image
    Given the database is empty
    And I am on the main menu
    When I type "docker build -t myapp:latest ."
    And I press enter
    Then I should see the build output

  Scenario: Navigate docker stats
    Given I am on the main menu
    When I type "docker stats"
    And I press enter
    And I press "j" to navigate down
    And I press "k" to navigate up
    And I cancel

  Scenario: Stop and remove a container
    Given I am on the main menu
    When I type "docker stop myapp"
    And I press enter
    And I type "docker rm myapp"
    And I press enter
    Then the container should be removed
