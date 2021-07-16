package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

// StateImportCompleted is the 'completed' state for an import job
const StateImportCompleted = "completed"

// MaxConflictRetries defines the maximum number of times that a post request will be retired after a conflict response
const MaxConflictRetries = 10

// ConflictRetryPeriod is the initial time period between post dimension option retries
var ConflictRetryPeriod = 250 * time.Millisecond

// CategoryDimensionImport is the handle for the CategoryDimensionImport event
type CategoryDimensionImport struct {
	cfg       config.Config
	ctblr     CantabularClient
	datasets  DatasetAPIClient
	importApi ImportAPIClient
	producer  kafka.IProducer
}

// NewCategoryDimensionImport creates a new CategoryDimensionImport with the provided config and clients.
// Note that for clients using the Client provided by dp-net/http, a mechanism to retry on 5xx status code is already in place.
func NewCategoryDimensionImport(cfg config.Config, c CantabularClient, d DatasetAPIClient, i ImportAPIClient, p kafka.IProducer) *CategoryDimensionImport {
	return &CategoryDimensionImport{
		cfg:       cfg,
		ctblr:     c,
		datasets:  d,
		importApi: i,
		producer:  p,
	}
}

// getSubmittedInstance gets an instance from Dataset API and validates that it is in submitted state.
// If the instance could not be obtained, we try to set it to 'failed' state, but if the state validation fails, we do not change the state.
func (h *CategoryDimensionImport) getSubmittedInstance(ctx context.Context, e *event.CategoryDimensionImport, ifMatch string) (i dataset.Instance, eTag string, err error) {

	// get instance
	i, eTag, err = h.datasets.GetInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID, ifMatch)
	if err != nil {
		// set instance state to failed because it could not be obtained and the import process will be aborted.
		// TODO we might want to retry this, once retries are implemented
		return i, "", h.instanceFailed(ctx, fmt.Errorf("error getting instance from dataset-api: %w", err), e)
	}

	// validate that instance is in 'submitted' state
	if i.State != dataset.StateSubmitted.String() {
		return i, "", &Error{
			err:     errors.New("instance is in wrong state, no more dimensions options will be imported"),
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
		// set instance state to failed because cantabular data could not be obtained and the import process will be aborted.
		// TODO we might want to retry this, once retries are implemented
		return h.instanceFailed(ctx, fmt.Errorf("error getting cantabular codebook: %w", err), e)
	}

	// validate that there is exactly one Codebook in the response
	if resp == nil || resp.Codebook == nil || len(resp.Codebook) != 1 {
		err := NewError(errors.New("unexpected response from Cantabular server"), log.Data{"response": resp})
		// set instance state to failed because cantabular response is invalid and the import process will be aborted.
		return h.instanceFailed(ctx, err, e)
	}

	variable := resp.Codebook[0]

	// Post a new dimension option for each item. If the eTag value changes from one call to another, validate the instance state again, and abort if it is not 'submitted'
	// TODO we will probably need to replace this Post with a batched Patch dimension with arrays of options (for performance reasons if we have lots of dimension options), similar to what we did in Filter API.
	attempt := 0
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
			// ErrInvalidDatasetAPIResponse covers the case where the dataset API responded with an unexpected Status Code.
			// If the status code was 409 Conflict, then it means that the instance changed since the last call.
			case *dataset.ErrInvalidDatasetAPIResponse:
				if errPost.Code() == http.StatusConflict {

					// check if we have already attemtped to post the instance more than MaxConflictRetries times
					if attempt >= MaxConflictRetries {
						return h.instanceFailed(ctx, fmt.Errorf("aborting import process after %d retries resulting in conflict on post dimension", MaxConflictRetries), e)
					}

					// sleep an exponential random time before retrying
					SleepRandom(attempt)

					// check that the instance is still in 'submitted' state
					_, eTag, err = h.getSubmittedInstance(ctx, e, headers.IfMatchAnyETag)
					if err != nil {
						return err
					}

					// instance is still in valid state and eTag has been updated. Retry this iteration
					i--
					attempt++
					continue

				} else {
					// any other unexpected status code results in the import process failing
					return h.instanceFailed(ctx, fmt.Errorf("error posting instance dimension option: %w", err), e)
				}
			default:
				// any other error type results in the import process failing
				return h.instanceFailed(ctx, fmt.Errorf("error posting instance dimension option: %w", err), e)
			}
		}
		attempt = 0
	}

	log.Info(ctx, "successfully posted all dimension options to dataset api", log.Data{
		"codes":     variable.Codes,
		"labels":    variable.Labels,
		"dimension": e.DimensionID,
	})

	// Count the total number of dimension options for this instance
	dims, newETag, err := h.datasets.GetInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, &dataset.QueryParams{Limit: 0}, headers.IfMatchAnyETag)
	if err != nil {
		// TODO we might want to retry this, once retries are implemented
		return h.instanceFailed(ctx, fmt.Errorf("error counting instance dimensions: %w", err), e)
	}

	// if all dimension options have been added to the instance and the instance did not change between the last update and the count,
	// then we might need to change states and trigger a kafka message
	if newETag == eTag && dims.TotalCount == resp.Dataset.Size {
		if err = h.onLastDimension(ctx, e, eTag); err != nil {
			return err
		}
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

// onLastDimension handles the case where all dimensions have been updated to the instance.
// If the eTag did not change since the last update, then we know that this consumer was the last one to update the instance. In that case:
// - Set instance to edition-confirmed
// - Set import job to completed state
// - send an InstanceComplete kafka message
func (h *CategoryDimensionImport) onLastDimension(ctx context.Context, e *event.CategoryDimensionImport, eTag string) error {

	// set instance to 'edition-confirmed' state, only if the eTag value did not change
	newETag, err := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateEditionConfirmed, eTag)
	if err != nil {
		switch putErr := err.(type) {
		case *dataset.ErrInvalidDatasetAPIResponse:
			// ErrInvalidDatasetAPIResponse covers the case where the dataset API responded with an unexpected Status Code.
			// If the status code was 409 Conflict, then it means that the instance changed since the last call.
			if putErr.Code() == http.StatusConflict {
				log.Info(ctx, "instance state could not be set to 'edition-confirmed' because the eTag value changed between posting the last dimension option and counting the number of dimension options. This is a valid scenario, another consumer was the last one to update the instance, and it will be responsibe of updating the instance and import job states, and sending the kafka message",
					log.Data{
						"event":    e,
						"old_etag": eTag,
						"new_etag": newETag,
					})
				return nil
			}
		}
		// for generic errors or any other unexpected status code, we return an error.
		return &Error{
			err:     fmt.Errorf("error while trying to set the instance to edition-confirmed state: %w", err),
			logData: log.Data{"event": e},
		}
	}

	// Import api update job to completed state
	if err := h.importApi.UpdateImportJobState(ctx, e.JobID, h.cfg.ServiceAuthToken, StateImportCompleted); err != nil {
		return fmt.Errorf("error updating import job to completed state: %w", err)
	}

	// create InstanceComplete event and Marshal it
	bytes, err := schema.InstanceComplete.Marshal(&event.InstanceComplete{
		InstanceID: e.InstanceID,
	})
	if err != nil {
		return err
	}

	// Send bytes to kafka producer output channel
	h.producer.Channels().Output <- bytes

	return nil
}

// getRetryTime will return a time based on the attempt and initial retry time.
// It uses the algorithm 2^n where n is the attempt number (double the previous) and
// a randomization factor of between 0-5ms so that the server isn't being hit constantly
// at the same time by many clients.
func getRetryTime(attempt int, retryTime time.Duration) time.Duration {
	n := (math.Pow(2, float64(attempt)))
	rand.Seed(time.Now().Unix())
	rnd := time.Duration(rand.Intn(4)+1) * time.Millisecond
	return (time.Duration(n) * retryTime) - rnd
}

// SleepRandom sleeps for a random period of time, determined by the provided attempt and the getRetryTime func
var SleepRandom = func(attempt int) {
	time.Sleep(getRetryTime(attempt, ConflictRetryPeriod))
}
