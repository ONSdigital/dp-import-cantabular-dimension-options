package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
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

// getSubmittedInstance gets an instance from Dataset API and validates that it is in submitted state.
// If the instance could not be obtained, we try to set it to 'failed' state
func (h *CategoryDimensionImport) getSubmittedInstance(ctx context.Context, e *event.CategoryDimensionImport, ifMatch string) (i dataset.Instance, eTag string, err error) {

	// get instance
	i, eTag, err = h.datasets.GetInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID, ifMatch)
	if err != nil {
		// TODO we might want to retry this, once retries are implemented
		return i, "", h.instanceFailed(ctx, fmt.Errorf("error getting instance from dataset-api: %w", err), e)
	}

	// validate instance state
	if i.State != dataset.StateSubmitted.String() {
		return i, "", &Error{
			err:     errors.New("instance is in wrong state, no dimensions options will be imported"),
			logData: log.Data{"event": e, "instance_state": i.State},
		}
	}

	return i, eTag, err
}

// Handle calls Cantabular server to obtain a list of variables for a CantabularBlob,
// which are then posted to the dataset API as dimension options
func (h *CategoryDimensionImport) Handle(ctx context.Context, e *event.CategoryDimensionImport) error {
	logData := log.Data{
		"event": e,
	}
	log.Info(ctx, "event handler called", logData)

	// get instance state and check that it is in submitted state
	_, eTag, err := h.getSubmittedInstance(ctx, e, headers.IfMatchAnyETag)
	if err != nil {
		return err
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
		err := NewError(errors.New("unexpected response from Cantabular server"), log.Data{"response": resp})
		return h.instanceFailed(ctx, err, e)
	}

	variable := resp.Codebook[0]

	// Post a new dimension option for each item. If the eTag value changes from one call to another, validate the instance state again, and abort if it is not 'submitted'
	// TODO we will probably need to replace this Post with a batched Patch dimension with arrays of options (for performance reasons if we have lots of dimeinsion options), similar to what we did in Filter API.
	for i := 0; i < variable.Len; i++ {
		_, err := h.datasets.PostInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.OptionPost{
			Name:     variable.Name,
			CodeList: variable.Name,     // TODO can we assume this?
			Code:     variable.Codes[i], // TODO can we assume this?
			Option:   variable.Codes[i],
			Label:    variable.Labels[i],
		}, eTag)
		if err != nil {
			switch errPost := err.(type) {
			case dataset.ErrInvalidDatasetAPIResponse:
				if errPost.Code() == http.StatusConflict {
					_, eTag, err = h.getSubmittedInstance(ctx, e, headers.IfMatchAnyETag)
					if err != nil {
						return err
					}
					i-- // retry with new eTag value, as the instance is still in a valid state
					continue
				}
			default:
				return h.instanceFailed(ctx, fmt.Errorf("error posting instance dimension option: %w", err), e)
			}
		}
	}

	log.Info(ctx, "successfully posted all dimension options to dataset api", log.Data{
		"codes":     variable.Codes,
		"labels":    variable.Labels,
		"dimension": e.DimensionID,
	})

	// Count the total number of dimension options for this instance, and update the instance state to 'completed' if all values have been imported
	dims, eTag, err := h.datasets.GetInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, &dataset.QueryParams{Limit: 0}, headers.IfMatchAnyETag)
	if err != nil {
		return h.instanceFailed(ctx, fmt.Errorf("error counting instance dimensions: %w", err), e)
	}

	if dims.Count == resp.Dataset.Size {
		// Check if this was the last dimension that was updated - then update instance to 'completed' state
		// Set instance to complete if and only if all dimension options have been added and we were the last ones to add a dimension option (i.e. eTag did not change between updating and counting)
		// This will guarantee that this is done exactly once, because if the instance changed, this request will fail with 409, and the other consumer that set the last value will do the update
		_, err = h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateCompleted, eTag)
		if err != nil {
			switch putErr := err.(type) {
			case dataset.ErrInvalidDatasetAPIResponse:
				if putErr.Code() == http.StatusConflict {
					log.Info(ctx, "instance state could not be set to 'completed' because it changed between putting the last dimension option and counting")
					return nil // valid scenario, another consumer was the last one to update the post instances.
				}
			}
			return &Error{
				err:     err,
				logData: log.Data{"event": e},
			}
		}

		// TODO import api update

		// TODO send kafka message
	}

	return nil
}

// instanceFailed updates the instance state to 'failed' and returns an Error wrapping the original error and any error during the state update
func (h *CategoryDimensionImport) instanceFailed(ctx context.Context, err error, e *event.CategoryDimensionImport) error {
	if _, err1 := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateFailed, headers.IfMatchAnyETag); err1 != nil {
		return &Error{
			err:     fmt.Errorf("error updating instance state during error handling: %w", err1),
			logData: log.Data{"event": e, "original_error": err},
		}
	}
	return err
}
