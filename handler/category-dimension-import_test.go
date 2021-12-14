package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler/mock"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errCantabular = errors.New("cantabular error")
	errDataset    = errors.New("dataset api error")
	errImportAPI  = errors.New("import api error")
	testToken     = "testToken"
	testCfg       = config.Config{
		ServiceAuthToken: testToken,
		BatchSizeLimit:   2,
	}
	testETag       = "testETag"
	newETag        = "newETag"
	testInstanceID = "test-instance-id"
	testJobID      = "test-job-id"
	testBlob       = "test-blob"
	ctx            = context.Background()
	testEvent      = event.CategoryDimensionImport{
		InstanceID:     testInstanceID,
		JobID:          testJobID,
		DimensionID:    "test-variable",
		CantabularBlob: testBlob,
	}
)

func TestHandle(t *testing.T) {

	Convey("Given a successful event handler, valid cantabular data, and an instance in completed state", t, func() {

		// mock SleepRandom to prevent delays in unit tests and to be able to validate the SleepRandom calls
		sleepRandomCalls := []int{}
		originalSleepRandom := handler.SleepRandom
		handler.SleepRandom = func(attempt int) {
			sleepRandomCalls = append(sleepRandomCalls, attempt)
		}
		defer func() {
			handler.SleepRandom = originalSleepRandom
		}()

		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		importAPIClient := importAPIClientHappy(false, false)
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

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

			Convey("Then 2 patch calls are performed to Dataset API, each containing a batch of Cantabular variable codes", func() {
				So(datasetAPIClient.PatchInstanceDimensionsCalls(), ShouldHaveLength, 2)

				// First batch has the first 2 items
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].Data, ShouldResemble, []*dataset.OptionPost{
					{
						Code:     "code1",
						Option:   "code1",
						Label:    "Code 1",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
					{
						Code:     "code2",
						Option:   "code2",
						Label:    "Code 2",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
				})

				// Second batch has the remaining item
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].Data, ShouldResemble, []*dataset.OptionPost{
					{
						Code:     "code3",
						Option:   "code3",
						Label:    "Code 3",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
				})
			})

			Convey("Then we do not sleep between calls", func() {
				So(sleepRandomCalls, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, an instance in completed state and that the last dimension of the whole import process has been imported by this consumer", t, func() {
		// mock SleepRandom to prevent delays in unit tests and to be able to validate the SleepRandom calls
		sleepRandomCalls := []int{}
		originalSleepRandom := handler.SleepRandom
		handler.SleepRandom = func(attempt int) {
			sleepRandomCalls = append(sleepRandomCalls, attempt)
		}
		defer func() {
			handler.SleepRandom = originalSleepRandom
		}()

		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyLastDimension()
		importAPIClient := importAPIClientHappy(true, true)
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
					InstanceID:     testInstanceID,
					CantabularBlob: testBlob,
				})
				So(err, ShouldBeNil)
				sentBytes := <-producer.Channels().Output
				So(sentBytes, ShouldResemble, expectedBytes)
			})

			Convey("Then we do not sleep between calls", func() {
				So(sleepRandomCalls, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, an instance in completed state and that the last dimension of the instance, but not the whole import process has been imported by this consumer", t, func() {
		// mock SleepRandom to prevent delays in unit tests and to be able to validate the SleepRandom calls
		sleepRandomCalls := []int{}
		originalSleepRandom := handler.SleepRandom
		handler.SleepRandom = func(attempt int) {
			sleepRandomCalls = append(sleepRandomCalls, attempt)
		}
		defer func() {
			handler.SleepRandom = originalSleepRandom
		}()

		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyLastDimension()
		importAPIClient := importAPIClientHappy(true, false)
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

			Convey("Then the import job is not updated", func() {
				So(importAPIClient.UpdateImportJobStateCalls(), ShouldHaveLength, 0)
			})

			Convey("Then the expected InstanceComplete event is sent to the kafka producer", func() {
				expectedBytes, err := schema.InstanceComplete.Marshal(&event.InstanceComplete{
					InstanceID:     testInstanceID,
					CantabularBlob: testBlob,
				})
				So(err, ShouldBeNil)
				sentBytes := <-producer.Channels().Output
				So(sentBytes, ShouldResemble, expectedBytes)
			})

			Convey("Then we do not sleep between calls", func() {
				So(sleepRandomCalls, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, an instance in completed state and that not all dimensions have been processed yet", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyLastDimension()
		importAPIClient := importAPIClientHappy(false, false)
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

		Convey("When Handle is triggered", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the handler does not try to update the instance state", func() {
				So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a successful event handler, valid cantabular data, and an instance in completed state, with an ETag that changes after the first post", t, func() {
		// mock SleepRandom to prevent delays in unit tests and to be able to validate the SleepRandom calls
		sleepRandomCalls := []int{}
		originalSleepRandom := handler.SleepRandom
		handler.SleepRandom = func(attempt int) {
			sleepRandomCalls = append(sleepRandomCalls, attempt)
		}
		defer func() {
			handler.SleepRandom = originalSleepRandom
		}()

		ctblrClient := cantabularClientHappy()
		datasetAPIClient := mock.DatasetAPIClientMock{}
		datasetAPIClient.PatchInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
			switch len(datasetAPIClient.PatchInstanceDimensionsCalls()) {
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
					State: dataset.StateCompleted.String(),
				},
			}
			switch len(datasetAPIClient.PatchInstanceDimensionsCalls()) {
			case 0, 1:
				return inst, testETag, nil
			default:
				return inst, newETag, nil
			}
		}
		importAPIClient := importAPIClientHappy(false, false)
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

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
				So(datasetAPIClient.PatchInstanceDimensionsCalls(), ShouldHaveLength, 3)

				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].IfMatch, ShouldEqual, testETag)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[0].Data, ShouldResemble, []*dataset.OptionPost{
					{
						Code:     "code1",
						Option:   "code1",
						Label:    "Code 1",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
					{
						Code:     "code2",
						Option:   "code2",
						Label:    "Code 2",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
				})

				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].IfMatch, ShouldEqual, testETag)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[1].Data, ShouldResemble, []*dataset.OptionPost{
					{
						Code:     "code3",
						Option:   "code3",
						Label:    "Code 3",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
				})

				So(datasetAPIClient.PatchInstanceDimensionsCalls()[2].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[2].ServiceAuthToken, ShouldEqual, testToken)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[2].IfMatch, ShouldEqual, newETag)
				So(datasetAPIClient.PatchInstanceDimensionsCalls()[2].Data, ShouldResemble, []*dataset.OptionPost{
					{
						Code:     "code3",
						Option:   "code3",
						Label:    "Code 3",
						CodeList: "test-variable",
						Name:     "test-variable",
					},
				})
			})

			Convey("Then we slept once between calls with an attempt value of 0", func() {
				So(sleepRandomCalls, ShouldHaveLength, 1)
				So(sleepRandomCalls[0], ShouldEqual, 0)
			})
		})
	})
}

func TestHandleFailure(t *testing.T) {

	Convey("Given a handler with a dataset api client that returns an instance in a non-completed state", t, func() {
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateFailed.String(),
					},
				}, testETag, nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, nil, &datasetAPIClient, nil, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, handler.NewError(
				errors.New("instance is in wrong state, no more dimensions options will be imported"),
				log.Data{"event": &testEvent, "instance_state": dataset.StateFailed.String()},
			))
		})
	})

	Convey("Given a handler with a cantabular client that returns an error", t, func() {
		ctblrClient := cantabularClientUnhappy()
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateCompleted.String(),
					},
				}, testETag, nil
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		importAPIClient := importAPIClientWithUpdateState()
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

		Convey("Then when Handle is triggered, the wrapped error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, fmt.Errorf("error getting cantabular codebook: %w", errCantabular))
			validateFailed(datasetAPIClient, importAPIClient)
		})

		Convey("And dataset API fails to set the instance state", func() {
			datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return "", errDataset
			}

			Convey("Then when Handle is triggered, the error with the event and nested error info is returned", func() {
				err := eventHandler.Handle(ctx, &testEvent)
				So(err, ShouldResemble, handler.NewError(
					fmt.Errorf("error getting cantabular codebook: %w", errCantabular),
					log.Data{
						"additional_errors": []error{
							fmt.Errorf("failed to update instance: %w", errDataset),
						},
					}),
				)
				validateFailed(datasetAPIClient, importAPIClient)
			})
		})
	})

	Convey("Given a handler with a cantabular client that returns an invalid response", t, func() {
		ctblrClient := cantabularInvalidResponse()
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateCompleted.String(),
					},
				}, testETag, nil
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return testETag, nil
			},
		}
		importAPIClient := importAPIClientWithUpdateState()
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

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
			validateFailed(datasetAPIClient, importAPIClient)
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
		importAPIClient := importAPIClientWithUpdateState()
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("error getting instance from dataset-api: %w", errDataset))
			})

			validateFailed(datasetAPIClient, importAPIClient)
		})
	})

	Convey("Given a handler with a dataset API and cantabular clients", t, func() {
		ctblrClient := cantabularClientHappy()
		importAPIClient := importAPIClientHappy(false, false)
		datasetAPIClient := mock.DatasetAPIClientMock{
			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				return dataset.Instance{
					Version: dataset.Version{
						State: dataset.StateCompleted.String(),
					},
				}, testETag, nil
			},
			PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
				return newETag, nil
			},
		}

		Convey("Where dataset API returns a 500 error on PostInstanceDimensions", func() {
			errPostInstance := dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusInternalServerError}, "uri")
			datasetAPIClient.PatchInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPostInstance
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

			Convey("Then when Handle is triggered", func() {
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected wrapped error is returned", func() {
					So(err, ShouldResemble,
						fmt.Errorf("failed to send dimension options to dataset api in batched patches: %w",
							fmt.Errorf("error processing a batch of cantabular variable values as dimension options: %w",
								fmt.Errorf("error patching instance dimensions: %w",
									errPostInstance,
								),
							),
						),
					)
				})

				validateFailed(datasetAPIClient, importAPIClient)
			})
		})

		Convey("Where dataset API returns a generic error on PatchInstanceDimensions", func() {
			errPatchInstance := errors.New("generic Dataset API Client Error")
			datasetAPIClient.PatchInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPatchInstance
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

			Convey("Then when Handle is triggered", func() {
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected error is returned", func() {
					So(err, ShouldResemble,
						fmt.Errorf("failed to send dimension options to dataset api in batched patches: %w",
							fmt.Errorf("error processing a batch of cantabular variable values as dimension options: %w",
								fmt.Errorf("error patching instance dimensions: %w",
									errPatchInstance,
								),
							),
						),
					)
				})

				validateFailed(datasetAPIClient, importAPIClient)
			})
		})

		Convey("Where dataset API returns a Conflict error on PostInstanceDimensions and the instance has changed to state failed", func() {
			errPostInstance := dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusConflict}, "uri")
			datasetAPIClient.PatchInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPostInstance
			}
			datasetAPIClient.GetInstanceFunc = func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
				switch len(datasetAPIClient.GetInstanceCalls()) {
				case 0:
					return dataset.Instance{
						Version: dataset.Version{
							State: dataset.StateCompleted.String(),
						},
					}, testETag, nil
				default:
					return dataset.Instance{
						Version: dataset.Version{
							State: dataset.StateFailed.String(),
						},
					}, newETag, nil
				}
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

			Convey("Then when Handle is triggered", func() {
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected error is returned", func() {
					So(err, ShouldResemble, handler.NewError(
						errors.New("instance is in wrong state, no more dimensions options will be imported"),
						log.Data{"event": &testEvent, "instance_state": dataset.StateFailed.String()},
					))
				})

				Convey("Then the instance state is not changed", func() {
					So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Where dataset API always returns a Conflict error on PostInstanceDimensions", func() {
			// mock SleepRandom to prevent delays in unit tests and to be able to validate the SleepRandom calls
			sleepRandomCalls := []int{}
			originalSleepRandom := handler.SleepRandom
			handler.SleepRandom = func(attempt int) {
				sleepRandomCalls = append(sleepRandomCalls, attempt)
			}
			defer func() {
				handler.SleepRandom = originalSleepRandom
			}()

			errPostInstance := dataset.NewDatasetAPIResponse(&http.Response{StatusCode: http.StatusConflict}, "uri")
			datasetAPIClient.PatchInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
				return "", errPostInstance
			}
			eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

			Convey("Then when Handle is triggered", func() {
				sleepRandomCalls = []int{}
				err := eventHandler.Handle(ctx, &testEvent)

				Convey("Then the expected error is returned", func() {
					So(err, ShouldResemble,
						fmt.Errorf("failed to send dimension options to dataset api in batched patches: %w",
							fmt.Errorf("error processing a batch of cantabular variable values as dimension options: %w",
								errors.New("aborting import process after 10 retries resulting in conflict on post dimension"),
							),
						),
					)
				})

				Convey("Then the post instance dimensions is retried MaxConflictRetries times", func() {
					So(datasetAPIClient.PatchInstanceDimensionsCalls(), ShouldHaveLength, handler.MaxConflictRetries+1)
				})

				Convey("Then the random sleep is called MaxConflictRetries times with the expected attemt values", func() {
					So(sleepRandomCalls, ShouldResemble, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
				})

				validateFailed(datasetAPIClient, importAPIClient)
			})
		})
	})

	Convey("Given a handler with an instance in completed state and that the last dimension has been imported by this consumer, but the state fails to change", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyLastDimension()
		importAPIClient := importAPIClientHappy(true, false)
		datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return "", fmt.Errorf("dataset api failed to set instance to %s state", state)
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, handler.NewError(
				handler.NewError(
					fmt.Errorf("error while trying to set the instance to edition-confirmed state: %w", errors.New("dataset api failed to set instance to edition-confirmed state")),
					log.Data{"event": &testEvent}),
				log.Data{
					"additional_errors": []error{
						fmt.Errorf(
							"failed to update instance: %w",
							errors.New("dataset api failed to set instance to failed state"),
						),
					},
				},
			))
		})
	})

	Convey("Given a handler with an import API client that retruns an error on IncreaseProcessedInstanceCount", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		datasetAPIClient.PutInstanceStateFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return newETag, nil
		}
		importAPIClient := mock.ImportAPIClientMock{
			IncreaseProcessedInstanceCountFunc: func(ctx context.Context, jobID string, serviceToken string, instanceID string) ([]importapi.ProcessedInstances, error) {
				return nil, errImportAPI
			},
			UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState string) error {
				return nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, nil)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("error increasing and counting instance count in import api: %w", errImportAPI))
			})

			validateFailed(datasetAPIClient, importAPIClient)
		})
	})

	Convey("Given a handler with with an import API that fails to ", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappyLastDimension()
		importAPIClient := importAPIClientHappy(true, true)
		importAPIClient.UpdateImportJobStateFunc = func(ctx context.Context, jobID string, serviceToken string, newState string) error {
			return errImportAPI
		}
		producer := kafkatest.NewMessageProducerWithChannels(&kafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}, true)

		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient, &importAPIClient, producer)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, fmt.Errorf("error updating import job to completed state: %w", errImportAPI))
			})

			Convey("Then the expected InstanceComplete event is sent to the kafka producer", func() {
				expectedBytes, err := schema.InstanceComplete.Marshal(&event.InstanceComplete{
					InstanceID:     testInstanceID,
					CantabularBlob: testBlob,
				})
				So(err, ShouldBeNil)
				sentBytes := <-producer.Channels().Output
				So(sentBytes, ShouldResemble, expectedBytes)
			})
		})
	})
}

func validateFailed(datasetAPIClient mock.DatasetAPIClientMock, importAPIClient mock.ImportAPIClientMock) {

	Convey("Then the instance is set to failed state in dataset API", func() {
		So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
		So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
		So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
		So(datasetAPIClient.PutInstanceStateCalls()[0].IfMatch, ShouldEqual, headers.IfMatchAnyETag)
	})

	Convey("Then the import job is set to failed state in import API", func() {
		So(importAPIClient.UpdateImportJobStateCalls(), ShouldHaveLength, 1)
		So(importAPIClient.UpdateImportJobStateCalls()[0].JobID, ShouldEqual, testJobID)
		So(importAPIClient.UpdateImportJobStateCalls()[0].NewState, ShouldEqual, handler.StateImportFailed)
	})
}

var testCodebookResp = &cantabular.GetCodebookResponse{
	Dataset: cantabular.Dataset{
		Size: 333,
	},
	Codebook: cantabular.Codebook{
		cantabular.Variable{
			VariableBase: cantabular.VariableBase{
				Name:  "test-variable",
				Label: "Test Variable",
			},
			Len: 3,
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

func cantabularClientHappy() mock.CantabularClientMock {
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return testCodebookResp, nil
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
		PatchInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
			return testETag, nil
		},
		GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
			return dataset.Instance{
				Version: dataset.Version{
					State: dataset.StateCompleted.String(),
				},
			}, testETag, nil
		},
	}
}

func datasetAPIClientHappyLastDimension() mock.DatasetAPIClientMock {
	return mock.DatasetAPIClientMock{
		PatchInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data []*dataset.OptionPost, ifMatch string) (string, error) {
			return testETag, nil
		},
		GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
			return dataset.Instance{
				Version: dataset.Version{
					State: dataset.StateCompleted.String(),
				},
			}, testETag, nil
		},
		PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
			return newETag, nil
		},
	}
}

func importAPIClientHappy(isLastInstanceDimension, isLastImportDimension bool) mock.ImportAPIClientMock {
	procInst := []importapi.ProcessedInstances{
		{
			ID:             testInstanceID,
			RequiredCount:  3,
			ProcessedCount: 2,
		},
		{
			ID:             "anotherInstance",
			RequiredCount:  5,
			ProcessedCount: 0,
		},
	}

	if isLastInstanceDimension {
		procInst[0].ProcessedCount = 3
	}

	if isLastImportDimension {
		procInst[1].ProcessedCount = 5
	}

	return mock.ImportAPIClientMock{
		IncreaseProcessedInstanceCountFunc: func(ctx context.Context, jobID string, serviceToken string, instanceID string) ([]importapi.ProcessedInstances, error) {
			return procInst, nil
		},
		UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState string) error {
			return nil
		},
	}
}

func importAPIClientWithUpdateState() mock.ImportAPIClientMock {
	return mock.ImportAPIClientMock{
		UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState string) error {
			return nil
		},
	}
}
