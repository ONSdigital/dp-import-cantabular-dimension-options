package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/rdumont/assistdog"
)

const testETag = "13c7791bafdbaaf5e6660754feb1a58cd6aaa892"

// DimensionCompleteEventsTimeout is the maximum time that the testing producer will wait
// for DimensionComplete events to be produced before failing the test
var DimensionCompleteEventsTimeout = time.Second

// RegisterSteps maps the human-readable regular expressions to their corresponding funcs
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following instance with id "([^"]*)" is available from dp-dataset-api:$`, c.theFollowingInstanceIsAvailable)
	ctx.Step(`^([^"]*) out of ([^"]*) dimensions have been processed for instance "([^"]*)" and job "([^"]*)"`, c.theCallToIncreaseProcessedInstanceIsSuccessful)
	ctx.Step(`^the following response is available from Cantabular from the codebook "([^"]*)" and query "([^"]*)":$`, c.theFollowingCodebookIsAvailable)

	ctx.Step(`^the call to add a dimension to the instance with id "([^"]*)" is successful`, c.theCallToAddInstanceDimensionIsSuccessful)
	ctx.Step(`^the instance with id "([^"]*)" is successfully updated`, c.theCallToUpdateInstanceIsSuccessful)
	ctx.Step(`^the job with id "([^"]*)" is successfully updated`, c.theCallToUpdateJobIsSuccessful)

	ctx.Step(`^this category-dimension-import event is consumed:$`, c.thisCategoryDimensionImportEventIsConsumed)
	ctx.Step(`^these instance-complete events are produced:$`, c.theseDimensionCompleteEventsAreProduced)
}

// theFollowingInstanceIsAvailable generate a mocked response for dataset API
// GET /instances/{id} with the provided instance response
func (c *Component) theFollowingInstanceIsAvailable(id string, instance *godog.DocString) error {
	c.DatasetAPI.NewHandler().
		Get("/instances/"+id).
		Reply(http.StatusOK).
		BodyString(instance.Content).
		AddHeader("Etag", testETag)

	return nil
}

// theFollowingCodebookIsAvailable generates a mocked response for Cantabular Server
// GET /v9/codebook/{name} with the provided query
func (c *Component) theFollowingCodebookIsAvailable(name, q string, cb *godog.DocString) error {
	c.CantabularSrv.NewHandler().
		Get("/v9/codebook/" + name + q).
		Reply(http.StatusOK).
		BodyString(cb.Content)

	return nil
}

// theCallToAddInstanceDimensionIsSuccessful generates a mocked response for Dataset API
// POST /instances/{id}/dimensions
func (c *Component) theCallToAddInstanceDimensionIsSuccessful(id string) error {
	c.DatasetAPI.NewHandler().
		Post("/instances/"+id+"/dimensions").
		Reply(http.StatusOK).
		AddHeader("ETag", testETag)

	return nil
}

// theCallToUpdateInstanceIsSuccessful generates a mocked response for Dataset API
// PUT /instances/{id}
func (c *Component) theCallToUpdateInstanceIsSuccessful(id string) error {
	c.DatasetAPI.NewHandler().
		Put("/instances/"+id).
		Reply(http.StatusOK).
		AddHeader("ETag", testETag)

	return nil
}

// theCallToUpdateJobIsSuccessful generates a mocked response for Import API
// PUT /jobs/{id}
func (c *Component) theCallToUpdateJobIsSuccessful(id string) error {
	c.ImportAPI.NewHandler().
		Put("/jobs/" + id).
		Reply(http.StatusOK)

	return nil
}

// theCallToIncreaseProcessedInstanceIsSuccessful generates a mocked response from Import API
// PUT /jobs/{jobID}/processed/{instanceID}
func (c *Component) theCallToIncreaseProcessedInstanceIsSuccessful(processed, required, instanceID, jobID string) error {

	// counts must be numbers
	processedCount, err := strconv.Atoi(processed)
	if err != nil {
		return fmt.Errorf("failed to convert processed to integer: %w", err)
	}
	requiredCount, err := strconv.Atoi(required)
	if err != nil {
		return fmt.Errorf("failed to convert required to integer: %w", err)
	}

	// response with the provided instance ID and counts
	resp := fmt.Sprintf(`[
		{
			"id": "%s",
			"required_count": %d,
			"processed_count": %d
		}
	]`, instanceID, processedCount, requiredCount)

	// create handler with the mocked response
	c.ImportAPI.NewHandler().
		Put(fmt.Sprintf("/jobs/%s/processed/%s", jobID, instanceID)).
		Reply(http.StatusOK).
		BodyString(resp)

	return nil
}

// theseDimensionCompleteEventsAreProduced consumes kafka messages that are expected to be produced by the service under test
// and validates that they match the expected values in the test
func (c *Component) theseDimensionCompleteEventsAreProduced(events *godog.Table) error {
	expected, err := assistdog.NewDefault().CreateSlice(new(event.InstanceComplete), events)
	if err != nil {
		return fmt.Errorf("failed to create slice from godog table: %w", err)
	}

	var got []*event.InstanceComplete
	listen := true

	for listen {
		select {
		case <-time.After(DimensionCompleteEventsTimeout):
			listen = false
		case <-c.consumer.Channels().Closer:
			return errors.New("closer channel closed")
		case msg, ok := <-c.consumer.Channels().Upstream:
			if !ok {
				return errors.New("upstream channel closed")
			}

			var e event.InstanceComplete
			var s = schema.InstanceComplete

			if err := s.Unmarshal(msg.GetData(), &e); err != nil {
				msg.Commit()
				msg.Release()
				return fmt.Errorf("error unmarshalling message: %w", err)
			}

			msg.Commit()
			msg.Release()

			got = append(got, &e)
		}
	}

	if diff := cmp.Diff(got, expected); diff != "" {
		return fmt.Errorf("-got +expected)\n%s\n", diff)
	}

	return nil
}

func (c *Component) thisCategoryDimensionImportEventIsConsumed(input *godog.DocString) error {
	ctx := context.Background()

	// testing kafka message that will be produced
	var testEvent event.CategoryDimensionImport
	if err := json.Unmarshal([]byte(input.Content), &testEvent); err != nil {
		return fmt.Errorf("error unmarshaling input to event: %w body: %s", err, input.Content)
	}

	log.Info(ctx, "event to marshal: ", log.Data{
		"event": testEvent,
	})

	// marshal and send message
	b, err := schema.CategoryDimensionImport.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event from schema: %w", err)
	}

	log.Info(ctx, "marshalled event: ", log.Data{
		"event": b,
	})

	c.producer.Channels().Output <- b

	return nil
}
