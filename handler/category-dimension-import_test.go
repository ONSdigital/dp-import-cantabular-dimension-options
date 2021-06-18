package handler

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
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
		eventHandler := &CategoryDimensionImport{
			cfg: config.Config{},
		}
		err := eventHandler.Handle(ctx, &testEvent)
		So(err, ShouldBeNil)

		// TODO add conveys to validate logic, when implemented
	})

	// TODO add negative tests cases when logic is implemented
}
