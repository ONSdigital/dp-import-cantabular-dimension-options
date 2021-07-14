package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler/mock"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errCantabular  = errors.New("cantabular error")
	errDataset     = errors.New("dataset api error")
	errImportAPI   = errors.New("import api error")
	testToken      = "testToken"
	testCfg        = config.Config{ServiceAuthToken: testToken}
	testETag       = "testETag"
	newETag        = "newETag"
	testInstanceID = "test-instance-id"
	testJobID      = "test-job-id"
	ctx            = context.Background()
	cantabularSize = 123
	testEvent      = event.CategoryDimensionImport{
		InstanceID:     testInstanceID,
		JobID:          testJobID,
		DimensionID:    "test-variable",
		CantabularBlob: "test-blob",
	}
)

func TestHandle(t *testing.T) {

	Convey("Given a successful event handler, valid cantabular data, and an instance in submitted state", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("When Handle is successfully triggered", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldBeNil)

			Convey("Then the corresponding codebook is obtained from cantabular", func() {
				So(ctblrClient.GetCodebookCalls(), ShouldHaveLength, 1)
				So(ctblrClient.GetCodebookCalls()[0].In2, ShouldResemble, cantabular.GetCodebookRequest{
					DatasetName: "test-blob",
					Categories:  true,
					Variables:   []string{"test-variable"},
				})
			})

			Convey("Then the corresponding instance is obtained from dataset API", func() {
				So(datasetAPIClient.GetInstanceCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.GetInstanceCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.GetInstanceCalls()[0].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
			})

			Convey("Then one Post call is performed to Dataset API for each Cantabular variable", func() {
				So(datasetAPIClient.PostInstanceDimensionsCalls(), ShouldHaveLength, 3)

				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code1",
					Option:   "code1",
					Label:    "Code 1",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code2",
					Option:   "code2",
					Label:    "Code 2",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code3",
					Option:   "code3",
					Label:    "Code 3",
					CodeList: "test-variable",
					Name:     "test-variable",
				})
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, an instance in submitted state and that the last dimension has been imported by this consumer", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyComplete()
		importAPIClient := mock.ImportAPIClientMock{
			UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState string) error {
				return nil
			},
		}
		producer := kafkatest.NewMessageProducerWithChannels(&kafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}, true)

		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, producer)

		Convey("When Handle is successfully triggered", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldBeNil)

			Convey("Then the instance is set to state completed", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldResemble, testInstanceID)
				So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldResemble, dataset.StateEditionConfirmed)
				So(datasetAPIClient.PutInstanceStateCalls()[0].IfMatch, ShouldResemble, testETag)
			})

			Convey("Then the import job is set to state completed", func() {
				So(importAPIClient.UpdateImportJobStateCalls(), ShouldHaveLength, 1)
				So(importAPIClient.UpdateImportJobStateCalls()[0].JobID, ShouldResemble, testJobID)
				So(importAPIClient.UpdateImportJobStateCalls()[0].NewState, ShouldResemble, handler.StateImportCompleted)
				So(importAPIClient.UpdateImportJobStateCalls()[0].ServiceToken, ShouldResemble, testToken)
			})

			Convey("Then the expected InstanceComplete event is sent to the kafka producer", func() {
				expectedBytes, err := schema.InstanceComplete.Marshal(&event.InstanceComplete{
					InstanceID: testInstanceID,
				})
				So(err, ShouldBeNil)
				sentBytes := <-producer.Channels().Output
				So(sentBytes, ShouldResemble, expectedBytes)
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, an instance in submitted state and that the last dimension has been imported by another consumer", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyComplete()
		datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return "", dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusConflict}, "uri")
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("When Handle is triggered", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the conflict error on update state is not returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the handler tries to set to state to edition-confirmed", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldResemble, testInstanceID)
				So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldResemble, dataset.StateEditionConfirmed)
				So(datasetAPIClient.PutInstanceStateCalls()[0].IfMatch, ShouldResemble, testETag)
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, and an instance in submitted state, with an ETag that changes after the first post", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := mock.DatasetAPIClientMock{}
		datasetAPIClient.PostInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
			switch len(datasetAPIClient.PostInstanceDimensionsCalls()) {
			case 0, 1:
				return testETag, nil
			case 2:
				return "", dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusConflict}, "uri")
			default:
				return newETag, nil
			}
		}
		datasetAPIClient.GetInstanceFunc = func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
			inst := dataset.Instance{
				Version: dataset.Version{
					State: dataset.StateSubmitted.String(),
				},
			}
			switch len(datasetAPIClient.PostInstanceDimensionsCalls()) {
			case 0, 1:
				return inst, testETag, nil
			default:
				return inst, newETag, nil
			}
		}
		datasetAPIClient.GetInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, q *dataset.QueryParams, ifMatch string) (dataset.Dimensions, string, error) {
			return testInstanceDimensions(3), newETag, nil
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("When Handle is successfully triggered", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldBeNil)

			Convey("Then the corresponding codebook is obtained from cantabular", func() {
				So(ctblrClient.GetCodebookCalls(), ShouldHaveLength, 1)
				So(ctblrClient.GetCodebookCalls()[0].In2, ShouldResemble, cantabular.GetCodebookRequest{
					DatasetName: "test-blob",
					Categories:  true,
					Variables:   []string{"test-variable"},
				})
			})

			Convey("Then the corresponding instance is obtained from dataset API twice, in order to validate the state initially and when the eTag changed", func() {
				So(datasetAPIClient.GetInstanceCalls(), ShouldHaveLength, 2)
				So(datasetAPIClient.GetInstanceCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.GetInstanceCalls()[0].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
				So(datasetAPIClient.GetInstanceCalls()[1].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.GetInstanceCalls()[1].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
			})

			Convey("Then one Post call is performed to Dataset API for each Cantabular variable, repeating the one that failed due to the eTag mismatch", func() {
				So(datasetAPIClient.PostInstanceDimensionsCalls(), ShouldHaveLength, 4)

				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].IfMatch, ShouldEqual, testETag)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code1",
					Option:   "code1",
					Label:    "Code 1",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].IfMatch, ShouldEqual, testETag)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code2",
					Option:   "code2",
					Label:    "Code 2",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].IfMatch, ShouldEqual, newETag)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code2",
					Option:   "code2",
					Label:    "Code 2",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[3].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[3].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[3].IfMatch, ShouldEqual, newETag)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[3].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code3",
					Option:   "code3",
					Label:    "Code 3",
					CodeList: "test-variable",
					Name:     "test-variable",
				})
			})
		})
	})
}

func TestHandleFailure(t *testing.T) {

	Convey("Given a handler with a dataset api client that returns an instance in a non-submitted state", t, func() {
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateCompleted.String(),
					},
				}, testETag, nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, nil, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, handler.NewError(
				errors.New("instance is in wrong state, no more dimensions options will be imported"),
				log.Data{"event": &testEvent, "instance_state": dataset.StateCompleted.String()},
			))
		})
	})

	Convey("Given a handler with a cantabular client that returns an error", t, func() {
		ctblrClient := cantabularClientUnhappy()
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateSubmitted.String(),
					},
				}, testETag, nil
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the wrapped error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, fmt.Errorf("error getting cantabular codebook: %w", errCantabular))

			Convey("Then the instance is set to failed state in dataset API", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
				So(datasetAPIClient.PutInstanceStateCalls()[0].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
			})
		})

		Convey("And dataset API fails to set the instance state", func() {
			datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return "", errDataset
			}

			Convey("Then when Handle is triggered, the error with the event and nested error info is returned", func() {
				err := eventHandler.Handle(ctx, &testEvent)
				So(err, ShouldResemble, handler.NewError(
					fmt.Errorf("error updating instance state during error handling: %w", errDataset),
					log.Data{
						"event":          &testEvent,
						"original_error": fmt.Errorf("error getting cantabular codebook: %w", errCantabular),
					},
				))

				Convey("Then the handler tires to set the instance to failed state in dataset API", func() {
					So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
					So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
					So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
					So(datasetAPIClient.PutInstanceStateCalls()[0].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
				})
			})
		})
	})

	Convey("Given a handler with a cantabular client that returns an invalid response", t, func() {
		ctblrClient := cantabularInvalidResponse()
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateSubmitted.String(),
					},
				}, testETag, nil
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the expected validation error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, handler.NewError(
				errors.New("unexpected response from Cantabular server"),
				log.Data{
					"response": &cantabular.GetCodebookResponse{
						Codebook: cantabular.Codebook{},
						Dataset:  cantabular.Dataset{},
					},
				}),
			)

			Convey("Then the instance is set to failed state in dataset API", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
				So(datasetAPIClient.PutInstanceStateCalls()[0].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
			})
		})
	})

	Convey("Given a handler with a dataset API client that returns an error on getInstance", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{}, "", errDataset
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return "", nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("error getting instance from dataset-api: %w", errDataset))
			})

			Convey("Then the instance state is set to failed", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
			})
		})
	})

	Convey("Given a handler with a dataset API and cantabular clients", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateSubmitted.String(),
					},
				}, testETag, nil
			},
			GetInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, q *dataset.QueryParams, ifMatch string) (dataset.Dimensions, string, error) {
				return testInstanceDimensions(3), testETag, nil
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return newETag, nil
			},
		}

		Convey("Where dataset API returns a 500 error on PostInstanceDimensions", func() {
			errPostInstance := dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusInternalServerError}, "uri")
			datasetAPIClient.PostInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPostInstance
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

			Convey("Then when Handle is triggered", func() {
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected error is returned", func() {
					So(err, ShouldResemble, fmt.Errorf("error posting instance dimension option: %w", errPostInstance))
				})

				Convey("Then the instance state is set to failed", func() {
					So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
					So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
					So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
				})
			})
		})

		Convey("Where dataset API returns a generic error on PostInstanceDimensions", func() {
			errPostInstance := errors.New("generic Dataset API Client Error")
			datasetAPIClient.PostInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPostInstance
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

			Convey("Then when Handle is triggered", func() {
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected error is returned", func() {
					So(err, ShouldResemble, fmt.Errorf("error posting instance dimension option: %w", errPostInstance))
				})

				Convey("Then the instance state is set to failed", func() {
					So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
					So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
					So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
				})
			})
		})

		Convey("Where dataset API returns a Conflict error on PostInstanceDimensions and the instance has changed to state completed", func() {
			errPostInstance := dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusConflict}, "uri")
			datasetAPIClient.PostInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPostInstance
			}
			datasetAPIClient.GetInstanceFunc = func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				switch len(datasetAPIClient.GetInstanceCalls()) {
				case 0:
					return dataset.Instance{
						Version: dataset.Version{
							State: dataset.StateSubmitted.String(),
						},
					}, testETag, nil
				default:
					return dataset.Instance{
						Version: dataset.Version{
							State: dataset.StateCompleted.String(),
						},
					}, newETag, nil
				}
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

			Convey("Then when Handle is triggered", func() {
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected error is returned", func() {
					So(err, ShouldResemble, handler.NewError(
						errors.New("instance is in wrong state, no more dimensions options will be imported"),
						log.Data{"event": &testEvent, "instance_state": dataset.StateCompleted.String()},
					))
				})

				Convey("Then the instance state is not changed", func() {
					So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 0)
				})
			})
		})
	})

	Convey("Given a handler with an instance in submitted state and that the last dimension has been imported by this consumer, but the state fail to change", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyComplete()
		datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return "", errDataset
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, handler.NewError(
				fmt.Errorf("error while trying to set the instance to edition-confirmed state: %w", errDataset),
				log.Data{"event": &testEvent},
			))
		})
	})

	Convey("Given a handler with a dataset API client that returns an error on getInstanceDimensions", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		datasetAPIClient.GetInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, q *dataset.QueryParams, ifMatch string) (dataset.Dimensions, string, error) {
			return dataset.Dimensions{}, "", errDataset
		}
		datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return newETag, nil
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("error counting instance dimensions: %w", errDataset))
			})

			Convey("Then the instance state is set to failed", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
				So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
			})
		})
	})

	Convey("Given a handler with with a failing import API", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyComplete()
		importAPIClient := mock.ImportAPIClientMock{
			UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState string) error {
				return errImportAPI
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("error updating import job to completed state: %w", errImportAPI))
			})
		})
	})
}

// testCodebookResp returns the expected Code
func testCodebookResp(totalSize int) *cantabular.GetCodebookResponse {
	return &cantabular.GetCodebookResponse{
		Dataset: cantabular.Dataset{
			Size: totalSize,
		},
		Codebook: cantabular.Codebook{
			cantabular.Variable{
				Name:  "test-variable",
				Label: "Test Variable",
				Len:   3,
				Codes: []string{
					"code1",
					"code2",
					"code3",
				},
				Labels: []string{
					"Code 1",
					"Code 2",
					"Code 3",
				},
			},
		},
	}
}

// testInstanceDimensions returns the expected response of a dataset API GET instance/dimensions call with limit=0 for a fully imported cantabular dataset
func testInstanceDimensions(totalCount int) dataset.Dimensions {
	return dataset.Dimensions{
		TotalCount: totalCount,
		Count:      0,
		Offset:     0,
		Limit:      0,
	}
}

func cantabularClientHappy() mock.CantabularClientMock {
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return testCodebookResp(cantabularSize), nil
		},
	}
}

func cantabularClientUnhappy() mock.CantabularClientMock {
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return nil, errCantabular
		},
	}
}

func cantabularInvalidResponse() mock.CantabularClientMock {
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return &cantabular.GetCodebookResponse{
				Codebook: cantabular.Codebook{},
			}, nil
		},
	}
}

func datasetAPIClientHappy() mock.DatasetAPIClientMock {
	return mock.DatasetAPIClientMock{
		PostInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
			return testETag, nil
		},
		GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
			return dataset.Instance{
				Version: dataset.Version{
					State: dataset.StateSubmitted.String(),
				},
			}, testETag, nil
		},
		GetInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, q *dataset.QueryParams, ifMatch string) (dataset.Dimensions, string, error) {
			return testInstanceDimensions(3), testETag, nil
		},
	}
}

func datasetAPIClientHappyComplete() mock.DatasetAPIClientMock {
	return mock.DatasetAPIClientMock{
		PostInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
			return testETag, nil
		},
		GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
			return dataset.Instance{
				Version: dataset.Version{
					State: dataset.StateSubmitted.String(),
				},
			}, testETag, nil
		},
		GetInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, q *dataset.QueryParams, ifMatch string) (dataset.Dimensions, string, error) {
			return testInstanceDimensions(cantabularSize), testETag, nil
		},
		PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return newETag, nil
		},
	}
}
