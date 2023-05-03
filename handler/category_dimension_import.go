package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
)

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

// getCompletedInstance gets an instance from Dataset API and validates that it is in completed state.
// If the instance could not be obtained, we try to set it to 'failed' state, but if the state validation fails, we do not change the state.
func (h *CategoryDimensionImport) getCompletedInstance(ctx context.Context, e *event.CategoryDimensionImport, ifMatch string) (i dataset.Instance, eTag string, err error) {
	// get instance
	i, eTag, err = h.datasets.GetInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID, ifMatch)
	if err != nil {
		// set instance state to failed because it could not be obtained and the import process will be aborted.
		// TODO we might want to retry this, once retries are implemented
		return i, "", h.setImportToFailed(ctx, fmt.Errorf("error getting instance from dataset-api: %w", err), e)
	}

	// validate that instance is in 'completed' state
	if i.State != dataset.StateCompleted.String() {
		return i, "", &Error{
			err:     errors.New("instance is in wrong state, no more dimensions options will be imported"),
			logData: log.Data{"event": e, "instance_state": i.State},
		}
	}

	return i, eTag, err
}

// Handle calls Cantabular server to obtain a list of variables for a CantabularBlob,
// which are then posted to the dataset API as dimension options
func (h *CategoryDimensionImport) Handle(ctx context.Context, workerID int, msg kafka.Message) error {
	e := &event.CategoryDimensionImport{}
	s := schema.CategoryDimensionImport

	if err := s.Unmarshal(msg.GetData(), e); err != nil {
		return h.setImportToFailed(ctx, fmt.Errorf("failed to unmarshal event: %w", err), e)
	}

	logData := log.Data{"event": e}
	log.Info(ctx, "event received", logData)

	// get instance state and check that it is in completed state
	instance, eTag, err := h.getCompletedInstance(ctx, e, headers.IfMatchAnyETag)
	if err != nil {
		return err
	}

	if !e.IsGeography {
		resp, err := h.ctblr.GetAggregatedDimensionOptions(ctx, cantabular.GetAggregatedDimensionOptionsRequest{
			Dataset:        e.CantabularBlob,
			DimensionNames: []string{e.DimensionID},
		})
		if err != nil {
			// TODO we might want to retry this, once retries are implemented
			return h.setImportToFailed(ctx, fmt.Errorf("error getting cantabular codebook: %w", err), e)
		}

		if resp == nil || len(resp.Dataset.Variables.Edges) != 1 {
			logData["response"] = resp
			err := NewError(errors.New("unexpected response from Cantabular server"), logData)
			return h.setImportToFailed(ctx, err, e)
		}

		// send variable values to dataset api in batches
		dim := resp.Dataset.Variables.Edges[0].Node
		eTag, err = h.BatchPatchInstance(ctx, e, &dim, instance, eTag)
		if err != nil {
			return fmt.Errorf("failed to send dimension options to dataset api in batched patches: %w", err)
		}
	}

	log.Info(ctx, "successfully sent all dimension options to dataset api for a dimension", logData)

	// Increase the import job with the instance counter and check if this was the last dimension for the instance
	procInst, err := h.importApi.IncreaseProcessedInstanceCount(ctx, e.JobID, h.cfg.ServiceAuthToken, e.InstanceID)
	if err != nil {
		// TODO we might want to retry this, once retries are implemented
		return h.setImportToFailed(ctx, fmt.Errorf("error increasing and counting instance count in import api: %w", err), e)
	}

	log.Info(ctx, "event processed (all dimensions for instance processed)- message will be committed by caller", logData)

	instanceLastDimension, importComplete := IsComplete(procInst, e.InstanceID)

	if instanceLastDimension {
		// set instance state to edition-confirmed and send kafka message
		if err = h.onLastDimension(ctx, e, eTag); err != nil {
			return h.setImportToFailed(ctx, err, e)
		}
		log.Info(ctx, "all dimensions in instance have been completely processed and kafka message has been sent", logData)
	}
	if importComplete {
		if err := h.importApi.UpdateImportJobState(ctx, e.JobID, h.cfg.ServiceAuthToken, importapi.StateCompleted); err != nil {
			return fmt.Errorf("error updating import job to completed state: %w", err)
		}
		log.Info(ctx, "all instances for the import have been successfully processed and the job has been set to completed state", logData)
	}

	return nil
}

// BatchPatchInstance sends new dimension options to Dataset API, corresponding to the provided Cantabular variable, in batches of up to BatchSizeLimit
func (h *CategoryDimensionImport) BatchPatchInstance(ctx context.Context, e *event.CategoryDimensionImport, dim *gql.Node, inst dataset.Instance, eTag string) (newETag string, err error) {
	numFullChunks := len(dim.Categories.Edges) / h.cfg.BatchSizeLimit
	remainingSize := len(dim.Categories.Edges) % h.cfg.BatchSizeLimit

	// Create a lookup of dimension ID's to names, so that we can
	// use a Cantabular name (e.g. `siblings_3`) to map to the name
	// stored against the dimension (e.g. `siblings`).
	idNameLookup := map[string]string{}
	for _, dimInst := range inst.Dimensions {
		idNameLookup[dimInst.ID] = dimInst.Name
	}

	// processBatch is a nested func to process a batch starting at the provided offset, with the provided size
	processBatch := func(offset, size int) {
		optionsBatch := make([]*dataset.OptionPost, size)
		for j := 0; j < size; j++ {
			optionsBatch[j] = &dataset.OptionPost{
				Name:     idNameLookup[strings.ToLower(dim.Name)],
				CodeList: dim.Name,                                 // TODO can we assume this?
				Code:     dim.Categories.Edges[offset+j].Node.Code, // TODO can we assume this?
				Option:   dim.Categories.Edges[offset+j].Node.Code,
				Label:    dim.Categories.Edges[offset+j].Node.Label,
			}
		}

		eTag, err = h.PatchInstanceDimensionsWithRetries(ctx, e, optionsBatch, eTag, 0)
		if err != nil {
			err = fmt.Errorf("error processing a batch of cantabular variable values as dimension options: %w", err)
		}
	}

	// process full batches
	for i := 0; i < numFullChunks; i++ {
		offset := i * h.cfg.BatchSizeLimit
		processBatch(offset, h.cfg.BatchSizeLimit)
		if err != nil {
			return "", err // err was wrapped by processBatch call
		}
	}

	// process any remaining options in a last smaller batch
	if remainingSize > 0 {
		offset := numFullChunks * h.cfg.BatchSizeLimit
		processBatch(offset, remainingSize)
		if err != nil {
			return "", err
		}
	}
	// some stuff

	return eTag, nil
}

// PatchInstanceDimensionsWithRetries sends a patch request to dataset API with the provided options.
// If the eTag value changes, validate the instance state again and retry only if it is still in 'completed' state.
// We will do up to MaxConflictRetries retires, if all of them fail, set the instance and import job to failed state
func (h *CategoryDimensionImport) PatchInstanceDimensionsWithRetries(ctx context.Context, e *event.CategoryDimensionImport, options []*dataset.OptionPost, eTag string, attempt int) (newETag string, err error) {
	eTag, err = h.datasets.PatchInstanceDimensions(ctx, h.cfg.ServiceAuthToken, e.InstanceID, options, nil, eTag)
	if err != nil {
		switch errPost := err.(type) {
		// ErrInvalidDatasetAPIResponse covers the case where the dataset API responded with an unexpected Status Code.
		// If the status code was 409 Conflict, then it means that the instance changed since the last call.
		case *dataset.ErrInvalidDatasetAPIResponse:
			if errPost.Code() == http.StatusConflict {
				// check if we have already attemtped to post the instance more than MaxConflictRetries times
				if attempt >= MaxConflictRetries {
					return "", h.setImportToFailed(ctx, fmt.Errorf("aborting import process after %d retries resulting in conflict on post dimension", MaxConflictRetries), e)
				}

				// sleep an exponential random time before retrying
				SleepRandom(attempt)

				// check that the instance is still in 'completed' state
				_, eTag, err = h.getCompletedInstance(ctx, e, headers.IfMatchAnyETag)
				if err != nil {
					return "", fmt.Errorf("error while patching dimensions, instance may be in a wrong state: %w", err)
				}

				// instance is still in valid state and eTag has been updated. Retry to process this batch.
				return h.PatchInstanceDimensionsWithRetries(ctx, e, options, eTag, attempt+1)

			} else {
				// any other unexpected status code results in the import process failing
				return "", h.setImportToFailed(ctx, fmt.Errorf("error patching instance dimensions: %w", err), e)
			}
		default:
			// any other error type results in the import process failing
			return "", h.setImportToFailed(ctx, fmt.Errorf("error patching instance dimensions: %w", err), e)
		}
	}
	return eTag, nil
}

// setImportToFailed updates the instance and the import states to 'failed' and returns an Error wrapping the original error and any other error during the state update calls
func (h *CategoryDimensionImport) setImportToFailed(ctx context.Context, err error, e *event.CategoryDimensionImport) error {
	additionalErrs := []error{}

	if _, errUpdateImport := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateFailed, headers.IfMatchAnyETag); errUpdateImport != nil {
		additionalErrs = append(additionalErrs, fmt.Errorf("failed to update instance: %w", errUpdateImport))
	}
	if errUpdateInstance := h.importApi.UpdateImportJobState(ctx, e.JobID, h.cfg.ServiceAuthToken, importapi.StateFailed); errUpdateInstance != nil {
		additionalErrs = append(additionalErrs, fmt.Errorf("failed to update import job state: %w", errUpdateInstance))
	}

	if len(additionalErrs) > 0 {
		return &Error{
			err:     err,
			logData: log.Data{"additional_errors": additionalErrs},
		}
	}

	return err
}

// IsComplete checks if the instance is complete and if all instances in import process are complete
func IsComplete(procInst []importapi.ProcessedInstances, instanceID string) (instanceLastDimension, importComplete bool) {
	importComplete = true
	instanceLastDimension = false
	for _, instCount := range procInst {
		if instCount.ProcessedCount != instCount.RequiredCount {
			importComplete = false
		} else {
			if instCount.ID == instanceID {
				instanceLastDimension = true
			}
		}
	}
	return instanceLastDimension, importComplete
}

// onLastDimension handles the case where all dimensions have been updated to the instance. The following actions will happen:
// - Set instance to edition-confirmed
// - send an InstanceComplete kafka message
func (h *CategoryDimensionImport) onLastDimension(ctx context.Context, e *event.CategoryDimensionImport, eTag string) error {
	// set instance to 'edition-confirmed' state, only if the eTag value did not change
	_, err := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateEditionConfirmed, eTag)
	if err != nil {
		return &Error{
			err:     fmt.Errorf("error while trying to set the instance to edition-confirmed state: %w", err),
			logData: log.Data{"event": e},
		}
	}

	// create InstanceComplete event and Marshal it
	bytes, err := schema.InstanceComplete.Marshal(&event.InstanceComplete{
		InstanceID:     e.InstanceID,
		CantabularBlob: e.CantabularBlob,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
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
