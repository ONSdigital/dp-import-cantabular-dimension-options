package event

import (
	"context"
	"os"
	"testing"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testEvent = HelloCalled{
		RecipientName: "World",
	}
)

// TODO: remove hello called example test
func TestHelloCalledHandler_Handle(t *testing.T) {
	ctx := context.Background()

	Convey("Given a successful event handler, when Handle is triggered", t, func() {
		filePath := "/tmp/helloworld.txt"
		eventHandler := &HelloCalledHandler{
			cfg: config.Config{OutputFilePath: filePath},
		}
		os.Remove(filePath)
		err := eventHandler.Handle(ctx, &testEvent)
		So(err, ShouldBeNil)
	})

	Convey("handler returns an error when cannot write to file", t, func() {
		filePath := ""
		eventHandler := &HelloCalledHandler{
			cfg: config.Config{OutputFilePath: filePath},
		}
		err := eventHandler.Handle(ctx, &testEvent)
		So(err, ShouldNotBeNil)
	})
}
