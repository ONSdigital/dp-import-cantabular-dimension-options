Feature: Import-Cantabular-Dimension-Options-Unhealthy

  Background:
    Given dp-dataset-api is unhealthy
    And dp-import-api is healthy
    And cantabular server is healthy
    And cantabular api extension is healthy
  
  Scenario: Not consuming category-dimension-import events, because a dependency is not healthy
    When the service starts
    And this category-dimension-import event is queued, to be consumed:
      """
      {
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "dimensionId":    "dimension-01",
        "CantabularBlob": "Example"
      }
      """  

    Then no instance-complete events should be produced
