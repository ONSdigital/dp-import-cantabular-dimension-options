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

	// get instance state
	i, err := h.datasets.GetInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID)
	if err != nil {
		// TODO we might want to retry this, once retries are implemented
		return h.instanceFailed(ctx, fmt.Errorf("error getting instance from dataset-api: %w", err), e)
	}

	// validate instance state
	if i.State != dataset.StateSubmitted.String() {
		return &Error{
			err:     errors.New("instance is in wrong state, no dimensions options will be imported"),
			logData: log.Data{"event": e, "instance_state": i.State},
		}
	}

	// obtain the possible values for the provided dimension and blob
	resp, err := h.ctblr.GetCodebook(ctx, cantabular.GetCodebookRequest{
		DatasetName: e.CantabularBlob,
		Categories:  true,
		Variables:   []string{e.DimensionID},
	})
	if err != nil {
		// TODO we might want to retry this, once retries are implemented
		return h.instanceFailed(ctx, fmt.Errorf("error getting cantabular codebook: %w", err), e)
	}

	// validate that there is exactly one Codebook in the response
	if resp == nil || resp.Codebook == nil || len(resp.Codebook) != 1 {
		return h.instanceFailed(ctx, errors.New("unexpected response from Cantabular server"), e)
	}

	variable := resp.Codebook[0]

	// Post a new dimension option for each item
	// TODO we will probably need to replace this Post with a batched Patch dimension with arrays of options (for performance reasons if we have lots of dimeinsion options), similar to what we did in Filter API.
	for i := 0; i < variable.Len; i++ {
		if err := h.datasets.PostInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.OptionPost{
			Name:     variable.Name,
			CodeList: variable.Name,     // TODO can we assume this?
			Code:     variable.Codes[i], // TODO can we assume this?
			Option:   variable.Codes[i],
			Label:    variable.Labels[i],
		}); err != nil {
			// TODO we might want to retry this, once retries are implemented
			return h.instanceFailed(ctx, fmt.Errorf("error posting instance dimension option: %w", err), e)
		}
	}

	// Count number of options for each dimension, if all dimensions have more than 0, atomically update the state to completed

	// check if this was the last dimension that was updated, according to the number of observations - then update instance and import task state to completed
	// TODO: we can have concurrency issues with this if multiple instances of this services handle different dimensions concurrently - we need an ETag - IfMatch mechanism in dataset API
	// it would be good enough if updateInstanceWithNewInserts returns the current value
	if err := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateCompleted); err != nil {
		return h.instanceFailed(ctx, fmt.Errorf("error updating instance to completed state: %w", err), e)
	}

	log.Info(ctx, "successfully posted all dimension options to dataset api", log.Data{
		"codes":     variable.Codes,
		"labels":    variable.Labels,
		"dimension": e.DimensionID,
	})

	return nil
}

// instanceFailed updates the instance state to 'failed' and returns an Error wrapping the original error and any error during the state update
func (h *CategoryDimensionImport) instanceFailed(ctx context.Context, err error, e *event.CategoryDimensionImport) error {
	if err1 := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateFailed); err1 != nil {
		return &Error{
			err:     fmt.Errorf("error updating instance state during error handling: %w", err1),
			logData: log.Data{"event": e, "original_error": err},
		}
	}
	return &Error{
		err:     err,
		logData: log.Data{"event": e},
	}
}
