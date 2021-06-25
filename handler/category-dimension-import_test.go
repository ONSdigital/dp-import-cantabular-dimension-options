package handler

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler/mock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testEvent = event.CategoryDimensionImport{
		JobID:       "testJobID",
		InstanceID:  "testInstanceID",
		DimensionID: "testDimensionID",
	}
)

func TestHandle(t *testing.T) {
	ctx := context.Background()

	Convey("Given a successful event handler, when Handle is triggered", t, func() {
		ctblrClient := cantabularClientHappy()

		eventHandler := &CategoryDimensionImport{
			cfg:   config.Config{},
			ctblr: &ctblrClient,
		}
		err := eventHandler.Handle(ctx, &testEvent)
		So(err, ShouldBeNil)

		// TODO add conveys to validate logic, when implemented
	})

	// TODO add negative tests cases when logic is implemented
}

func testCodebook() cantabular.Codebook {
	return cantabular.Codebook{}
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
