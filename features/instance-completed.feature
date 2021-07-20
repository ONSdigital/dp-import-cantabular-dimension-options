Feature: Import-Cantabular-Dimension-Options

  Background:
    Given the following response is available from Cantabular from the codebook "Example" and query "?cats=true&v=dimension-01":
      """
      {
        "dataset": {
          "name": "Example",
          "size": 5
        },
        "codebook": [
          {
            "name": "dimension-01",
            "label": "Dimension 01",
            "len": 3,
            "codes": [
              "0",
              "1",
              "2"
            ],
            "labels": [
              "London",
              "Liverpool",
              "Belfast"
            ]
          }
        ]
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
    When this category-dimension-import event is consumed:
      """
      {
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "dimensionId":    "dimension-01",
        "CantabularBlob": "Example"
      }
      """  
    And the call to add a dimension to the instance with id "instance-happy-01" is successful
    And the instance with id "instance-happy-01" is successfully updated to "edition-confirmed" state
    And the job with id "job-happy-01" is successfully updated to "completed" state
    And the instance with id "instance-happy-01" has 5 dimension options

    Then these instance-complete events are produced:
      | InstanceID        |
      | instance-happy-01 |
