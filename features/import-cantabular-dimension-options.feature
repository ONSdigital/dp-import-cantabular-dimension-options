Feature: Import-Cantabular-Dimension-Options

  Background:
    Given dp-dataset-api is healthy
    And dp-import-api is healthy
    And cantabular server is healthy
    And cantabular api extension is healthy
    And the following categories query response is available from Cantabular api extension for the dataset "Example" and variable "dimension-01":
      """
      {
        "data": {
          "dataset": {
            "table": {
              "dimensions": [
                {
                  "categories": [
                    {
                      "code": "0",
                      "label": "London"
                    },
                    {
                      "code": "1",
                      "label": "Liverpool"
                    },
                    {
                      "code": "2",
                      "label": "Belfast"
                    }
                  ],
                  "variable": {
                    "label": "Dimension 01",
                    "name": "dimension-01"
                  }
                }
              ]
            }
          }
        }
      }
      """
    And the following instance with id "instance-happy-01" is available from dp-dataset-api:
      """
      {
        "import_tasks": {
          "build_hierarchies": null,
          "build_search_indexes": null,
          "import_observations": {
            "total_inserted_observations": 0,
            "state": "created"
          }
        },
        "id": "057cd26b-e0ae-431f-9316-913db61cec39",
        "last_updated": "2021-07-19T09:59:28.417Z",
        "links": {
          "dataset": {
            "href": "http://localhost:22000/datasets/cantabular-dataset",
            "id": "cantabular-dataset"
          },
          "job": {
            "href": "http://localhost:21800/jobs/e7f99293-44f2-47ce-b6cb-db2f6618ef40",
            "id": "e7f99293-44f2-47ce-b6cb-db2f6618ef40"
          },
          "self": {
            "href": "http://10.201.4.160:10400/instances/057cd26b-e0ae-431f-9316-913db61cec39"
          }
        },
        "state": "completed"
      }
      """

  Scenario: Consuming a category-dimension-import event with correct fields
    When this category-dimension-import event is queued, to be consumed:
      """
      {
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "dimensionId":    "dimension-01",
        "CantabularBlob": "Example"
      }
      """  
    And the call to add a dimension to the instance with id "instance-happy-01" is successful
    And the instance with id "instance-happy-01" is successfully updated
    And the job with id "job-happy-01" is successfully updated
    And 2 out of 2 dimensions have been processed for instance "instance-happy-01" and job "job-happy-01"
    And the service starts

    Then these instance-complete events are produced:
      | InstanceID        | CantabularBlob |
      | instance-happy-01 | Example        |
      | instance-happy-01 | Example        |

  Scenario: Consuming a category-dimension-import event with correct fields but dimensions cannot be added
    When this category-dimension-import event is queued, to be consumed:
      """
      {
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "dimensionId":    "dimension-01",
        "CantabularBlob": "Example"
      }
      """  
    And the call to add a dimension to the instance with id "instance-happy-01" is unsuccessful
    And the instance with id "instance-happy-01" is successfully updated
    And the job with id "job-happy-01" is successfully updated
    And 2 out of 2 dimensions have been processed for instance "instance-happy-01" and job "job-happy-01"
    And the service starts

    Then no instance-complete events should be produced
