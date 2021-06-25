package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

// CategoryDimensionImport is the handle for the CategoryDimensionImport event
type CategoryDimensionImport struct {
	cfg      config.Config
	ctblr    CantabularClient
	datasets DatasetAPIClient
	producer kafka.IProducer
}

// NewCategoryDimensionImport creates a new CategoryDimensionImport with the provided config and Cantabular client
func NewCategoryDimensionImport(cfg config.Config, c CantabularClient, d DatasetAPIClient) *CategoryDimensionImport {
	return &CategoryDimensionImport{
		cfg:      cfg,
		ctblr:    c,
		datasets: d,
	}
}

// Handle calls Cantabular server to obtain a list of variables for a CantabularBlob,
// which are then posted to the dataset API as dimension options
func (h *CategoryDimensionImport) Handle(ctx context.Context, e *event.CategoryDimensionImport) error {
	logData := log.Data{
		"event": e,
	}
	log.Info(ctx, "event handler called", logData)

	resp, err := h.ctblr.GetCodebook(ctx, cantabular.GetCodebookRequest{
		DatasetName: e.CantabularBlob,
		Categories:  true,
		Variables:   []string{e.DimensionID},
	})
	if err != nil {
		return err
	}

	// validate that there is exactly one Codebook in the response
	if resp == nil || resp.Codebook == nil || len(resp.Codebook) != 1 {
		return errors.New("unexpected response from Cantabular server")
	}

	// Post a new dimension option for each item
	// TODO we might consider doing some posts concurrently if we observe bad performance
	for i := 0; i < resp.Codebook[0].Len; i++ {
		if err := h.datasets.PostInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.OptionPost{
			Name:     resp.Codebook[0].Name,
			CodeList: resp.Codebook[0].Name,     // TODO can we assume this?
			Code:     resp.Codebook[0].Codes[i], // TODO can we assume this?
			Option:   resp.Codebook[0].Codes[i],
			Label:    resp.Codebook[0].Labels[i],
		}); err != nil {
			return fmt.Errorf("error posting instance option: %w", err)
		}
	}

	// TODO call import API to update the state (or should this be done by the import tracker?)

	// TODO figure out how to know if all dimensions have been processed and send `import complete` kafka message
	// (or should this be done by the import tracker?)

	return nil
}
