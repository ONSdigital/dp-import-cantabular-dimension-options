package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler/mock"
	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errCantabular = errors.New("cantabular error")
	errDataset    = errors.New("dataset api error")
	testCfg       = config.Config{
		ServiceAuthToken: "testToken",
	}
)

func TestNewCategoryDimensionImport(t *testing.T) {

	Convey("NewCategoryDimensionImport returns the correct CategoryDimensionImport according to the provided params", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		created := NewCategoryDimensionImport(testCfg, &ctblrClient, &datasetAPIClient)
		So(created, ShouldResemble, &CategoryDimensionImport{
			cfg:      testCfg,
			ctblr:    &ctblrClient,
			datasets: &datasetAPIClient,
		})
	})

}

func TestHandle(t *testing.T) {
	ctx := context.Background()

	Convey("Given a successful event handler", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientHappy()

		eventHandler := &CategoryDimensionImport{
			cfg:      testCfg,
			ctblr:    &ctblrClient,
			datasets: &datasetAPIClient,
		}

		Convey("Then when Handle is triggered, one Post call is performed to Dataset API for each Cantabular variable", func() {
			err := eventHandler.Handle(ctx, &event.CategoryDimensionImport{
				InstanceID:     "test-instance-id",
				JobID:          "test-job-id",
				DimensionID:    "test-variable",
				CantabularBlob: "test-blob",
			})
			So(err, ShouldBeNil)

			So(ctblrClient.GetCodebookCalls(), ShouldHaveLength, 1)
			So(ctblrClient.GetCodebookCalls()[0].In2, ShouldResemble, cantabular.GetCodebookRequest{
				DatasetName: "test-blob",
				Categories:  true,
				Variables:   []string{"test-variable"},
			})

			So(datasetAPIClient.PostInstanceDimensionsCalls(), ShouldHaveLength, 3)

			So(datasetAPIClient.PostInstanceDimensionsCalls()[0].InstanceID, ShouldEqual, "test-instance-id")
			So(datasetAPIClient.PostInstanceDimensionsCalls()[0].ServiceAuthToken, ShouldEqual, "testToken")
			So(datasetAPIClient.PostInstanceDimensionsCalls()[0].Data, ShouldResemble, dataset.OptionPost{
				Code:     "code1",
				Option:   "code1",
				Label:    "Code 1",
				CodeList: "test-variable",
				Name:     "test-variable",
			})

			So(datasetAPIClient.PostInstanceDimensionsCalls()[1].InstanceID, ShouldEqual, "test-instance-id")
			So(datasetAPIClient.PostInstanceDimensionsCalls()[1].ServiceAuthToken, ShouldEqual, "testToken")
			So(datasetAPIClient.PostInstanceDimensionsCalls()[1].Data, ShouldResemble, dataset.OptionPost{
				Code:     "code2",
				Option:   "code2",
				Label:    "Code 2",
				CodeList: "test-variable",
				Name:     "test-variable",
			})

			So(datasetAPIClient.PostInstanceDimensionsCalls()[2].InstanceID, ShouldEqual, "test-instance-id")
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

	Convey("Given a handler with a cantabular client that returns an error", t, func() {
		ctblrClient := cantabularClientUnhappy()

		eventHandler := &CategoryDimensionImport{
			cfg:   testCfg,
			ctblr: &ctblrClient,
		}

		Convey("Then when Handle is triggered, the same error is returned", func() {
			err := eventHandler.Handle(ctx, &event.CategoryDimensionImport{
				InstanceID:     "test-instance-id",
				JobID:          "test-job-id",
				DimensionID:    "test-variable",
				CantabularBlob: "test-blob",
			})
			So(err, ShouldResemble, fmt.Errorf("error getting cantabular codebook: %w", errCantabular))
		})
	})

	Convey("Given a handler with a cantabular client that returns an invalid response", t, func() {
		ctblrClient := cantabularInvalidResponse()

		eventHandler := &CategoryDimensionImport{
			cfg:   testCfg,
			ctblr: &ctblrClient,
		}

		Convey("Then when Handle is triggered, the expected validation error is returned", func() {
			err := eventHandler.Handle(ctx, &event.CategoryDimensionImport{
				InstanceID:     "test-instance-id",
				JobID:          "test-job-id",
				DimensionID:    "test-variable",
				CantabularBlob: "test-blob",
			})
			So(err, ShouldResemble, &Error{
				err: errors.New("unexpected response from Cantabular server"),
				logData: log.Data{
					"response": &cantabular.GetCodebookResponse{
						Codebook: cantabular.Codebook{},
						Dataset:  cantabular.Dataset{},
					},
				},
			})
		})
	})

	Convey("Given a handler with a dataset API client that returns an error", t, func() {
		ctblrClient := cantabularClientHappy()
		datasetAPIClient := datasetAPIClientUnhappy()

		eventHandler := &CategoryDimensionImport{
			cfg:      testCfg,
			ctblr:    &ctblrClient,
			datasets: &datasetAPIClient,
		}

		Convey("Then when Handle is triggered, the expected error is returned", func() {
			err := eventHandler.Handle(ctx, &event.CategoryDimensionImport{
				InstanceID:     "test-instance-id",
				JobID:          "test-job-id",
				DimensionID:    "test-variable",
				CantabularBlob: "test-blob",
			})
			So(err, ShouldResemble, &Error{
				err:     fmt.Errorf("error posting instance option: %w", errDataset),
				logData: log.Data{"dimension": "test-variable"},
			})
		})
	})
}

func testCodebook() cantabular.Codebook {
	return cantabular.Codebook{
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
	}
}

func cantabularClientHappy() mock.CantabularClientMock {
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return &cantabular.GetCodebookResponse{
				Codebook: testCodebook(),
			}, nil
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
		PostInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost) error {
			return nil
		},
	}
}

func datasetAPIClientUnhappy() mock.DatasetAPIClientMock {
	return mock.DatasetAPIClientMock{
		PostInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost) error {
			return errDataset
		},
	}
}
