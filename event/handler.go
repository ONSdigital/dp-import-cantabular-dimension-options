package event

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// TODO: remove hello called example handler
// HelloCalledHandler ...
type HelloCalledHandler struct {
	cfg config.Config
}

// Handle takes a single event.
func (h *HelloCalledHandler) Handle(ctx context.Context, event *HelloCalled) (err error) {
	logData := log.Data{
		"event": event,
	}
	log.Info(ctx, "event handler called", logData)

	greeting := fmt.Sprintf("Hello, %s!", event.RecipientName)
	err = ioutil.WriteFile(h.cfg.OutputFilePath, []byte(greeting), 0644)
	if err != nil {
		return err
	}

	logData["greeting"] = greeting
	log.Event(ctx, "event successfully handled", log.INFO, logData)

	return nil
}
