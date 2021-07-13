package handler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler/mock"
	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errCantabular  = errors.New("cantabular error")
	errDataset     = errors.New("dataset api error")
	testCfg        = config.Config{ServiceAuthToken: "testToken"}
	testETag       = "testETag"
	testInstanceID = "test-instance-id"
	ctx            = context.Background()
	cantabularSize = 123
	testEvent      = event.CategoryDimensionImport{
		InstanceID:     testInstanceID,
		JobID:          "test-job-id",
		DimensionID:    "test-variable",
		CantabularBlob: "test-blob",
	}
)

func TestNewCategoryDimensionImport(t *testing.T) {

	Convey("NewCategoryDimensionImport returns the correct CategoryDimensionImport according to the provided params", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		created := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)
		So(created, ShouldResemble, handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient))
	})
}

func TestHandle(t *testing.T) {

	Convey("Given a successful event handler, valid cantabular data, and an instance in submitted state", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := mock.DatasetAPIClientMock{
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
				return testInstanceDimensions(), testETag, nil
			},
		}
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)

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
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].ServiceAuthToken, ShouldEqual, "testToken")
				So(datasetAPIClient.PostInstanceDimensionsCalls()[0].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code1",
					Option:   "code1",
					Label:    "Code 1",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].ServiceAuthToken, ShouldEqual, "testToken")
				So(datasetAPIClient.PostInstanceDimensionsCalls()[1].Data, ShouldResemble, dataset.OptionPost{
					Code:     "code2",
					Option:   "code2",
					Label:    "Code 2",
					CodeList: "test-variable",
					Name:     "test-variable",
				})

				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].InstanceID, ShouldEqual, testInstanceID)
				So(datasetAPIClient.PostInstanceDimensionsCalls()[2].ServiceAuthToken, ShouldEqual, "testToken")
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
		eventHandler := handler.NewCategoryDimensionImport(testCfg, nil, &datasetAPIClient)

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &testEvent)
			So(err, ShouldResemble, handler.NewError(
				errors.New("instance is in wrong state, no dimensions options will be imported"),
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
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)

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
		eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)

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

	// Convey("Given a handler with a dataset API client that returns an error on getInstance", t, func() {
	// 	ctblrClient := cantabularClientHappy()
	// 	datasetAPIClient := mock.DatasetAPIClientMock{
	// 		GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
	// 			return dataset.Instance{}, "", errDataset
	// 		},
	// 		PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
	// 			return "", nil
	// 		},
	// 	}
	// 	eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)

	// 	Convey("Then when Handle is triggered, the expected error is returned", func() {
	// 		err := eventHandler.Handle(ctx, &event.CategoryDimensionImport{
	// 			InstanceID:     testInstanceID,
	// 			JobID:          "test-job-id",
	// 			DimensionID:    "test-variable",
	// 			CantabularBlob: "test-blob",
	// 		})

	// 		Convey("Then the expected error is returned", func() {
	// 			So(err, ShouldResemble, handler.NewError(
	// 				fmt.Errorf("error posting instance option: %w", errDataset),
	// 				log.Data{"dimension": "test-variable"},
	// 			))
	// 		})

	// 		Convey("Then the instance state is set to failed", func() {
	// 			So(datasetAPIClient.PutInstanceStateCalls(), ShouldHaveLength, 1)
	// 			So(datasetAPIClient.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, testInstanceID)
	// 			So(datasetAPIClient.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed.String())
	// 		})
	// 	})
	// })

	// Convey("Given a handler with a dataset API client that returns an error on getInstance", t, func() {
	// 	ctblrClient := cantabularClientHappy()
	// 	datasetAPIClient := datasetAPIClientHappy()
	// 	datasetAPIClient.PostInstanceDimensionsFunc = func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
	// 		return "", errDataset
	// 	}
	// 	eventHandler := handler.NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)

	// 	Convey("Then when Handle is triggered, the expected error is returned", func() {
	// 		err := eventHandler.Handle(ctx, &event.CategoryDimensionImport{
	// 			InstanceID:     testInstanceID,
	// 			JobID:          "test-job-id",
	// 			DimensionID:    "test-variable",
	// 			CantabularBlob: "test-blob",
	// 		})
	// 		So(err, ShouldResemble, handler.NewError(
	// 			fmt.Errorf("error posting instance option: %w", errDataset),
	// 			log.Data{"dimension": "test-variable"},
	// 		))
	// 	})
	// })
}

// testCodebookResp returns the expected Code
func testCodebookResp() *cantabular.GetCodebookResponse {
	return &cantabular.GetCodebookResponse{
		Dataset: cantabular.Dataset{
			Size: cantabularSize,
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
func testInstanceDimensions() dataset.Dimensions {
	return dataset.Dimensions{
		TotalCount: cantabularSize,
		Count:      0,
		Offset:     0,
		Limit:      0,
	}
}

func cantabularClientHappy() mock.CantabularClientMock {
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return testCodebookResp(), nil
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
			return testInstanceDimensions(), testETag, nil
		},
	}
}
