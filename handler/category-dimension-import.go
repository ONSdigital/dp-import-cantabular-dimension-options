package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/log.go/v2/log"
)

// CategoryDimensionImport is the handle for the CategoryDimensionImport event
type CategoryDimensionImport struct {
	cfg      config.Config
	ctblr    CantabularClient
	datasets DatasetAPIClient
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

	// obtain the possible values for the provided dimension and blob
	resp, err := h.ctblr.GetCodebook(ctx, cantabular.GetCodebookRequest{
		DatasetName: e.CantabularBlob,
		Categories:  true,
		Variables:   []string{e.DimensionID},
	})
	if err != nil {
		return fmt.Errorf("error getting cantabular codebook: %w", err)
	}

	// validate that there is exactly one Codebook in the response
	if resp == nil || resp.Codebook == nil || len(resp.Codebook) != 1 {
		return &Error{
			err:     errors.New("unexpected response from Cantabular server"),
			logData: log.Data{"response": resp},
		}
	}

	variable := resp.Codebook[0]

	// Post a new dimension option for each item
	// TODO we might consider doing some posts concurrently if we observe bad performance
	for i := 0; i < variable.Len; i++ {
		if err := h.datasets.PostInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.OptionPost{
			Name:     variable.Name,
			CodeList: variable.Name,     // TODO can we assume this?
			Code:     variable.Codes[i], // TODO can we assume this?
			Option:   variable.Codes[i],
			Label:    variable.Labels[i],
		}); err != nil {
			return &Error{
				err:     fmt.Errorf("error posting instance option: %w", err),
				logData: log.Data{"dimension": e.DimensionID},
			}
		}
	}

	log.Info(ctx, "successfully posted all dimension options to dataset api", log.Data{
		"codes":  variable.Codes,
		"labels": variable.Labels,
	})

	return nil
}
