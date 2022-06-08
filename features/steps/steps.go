package steps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/rdumont/assistdog"
)

const testETag = "13c7791bafdbaaf5e6660754feb1a58cd6aaa892"

// RegisterSteps maps the human-readable regular expressions to their corresponding funcs
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following instance with id "([^"]*)" is available from dp-dataset-api:$`, c.theFollowingInstanceIsAvailable)
	ctx.Step(`^([^"]*) out of ([^"]*) dimensions have been processed for instance "([^"]*)" and job "([^"]*)"`, c.theCallToIncreaseProcessedInstanceIsSuccessful)
	ctx.Step(`^the following categories query response is available from Cantabular api extension for the dataset "([^"]*)" and variable "([^"]*)":$`, c.theFollowingCantabularCategoriesAreAvailable)

	ctx.Step(`^the service starts`, c.theServiceStarts)
	ctx.Step(`^dp-dataset-api is healthy`, c.datasetAPIIsHealthy)
	ctx.Step(`^dp-dataset-api is unhealthy`, c.datasetAPIIsUnhealthy)
	ctx.Step(`^dp-import-api is healthy`, c.importAPIIsHealthy)
	ctx.Step(`^cantabular server is healthy`, c.cantabularServerIsHealthy)
	ctx.Step(`^cantabular api extension is healthy`, c.cantabularAPIExtIsHealthy)

	ctx.Step(`^the call to add a dimension to the instance with id "([^"]*)" is successful`, c.theCallToAddInstanceDimensionIsSuccessful)
	ctx.Step(`^the call to add a dimension to the instance with id "([^"]*)" is unsuccessful`, c.theCallToAddInstanceDimensionIsUnsuccessful)
	ctx.Step(`^the instance with id "([^"]*)" is successfully updated`, c.theCallToUpdateInstanceIsSuccessful)
	ctx.Step(`^the job with id "([^"]*)" is successfully updated`, c.theCallToUpdateJobIsSuccessful)

	ctx.Step(`^this category-dimension-import event is queued, to be consumed:$`, c.thisCategoryDimensionImportEventIsQueued)
	ctx.Step(`^these instance-complete events are produced:$`, c.theseInstanceCompleteEventsAreProduced)
	ctx.Step(`^no instance-complete events should be produced`, c.noInstanceCompleteEventsShouldBeProduced)
}

// theServiceStarts starts the service under test in a new go-routine
// note that this step should be called only after all dependencies have been setup,
// to prevent any race condition, specially during the first healthcheck iteration.
func (c *Component) theServiceStarts() error {
	return c.startService(c.ctx)
}

// datasetAPIIsHealthy generates a mocked healthy response for dataset API healthecheck
func (c *Component) datasetAPIIsHealthy() error {
	const res = `{"status": "OK"}`
	c.DatasetAPI.NewHandler().
		Get("/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// datasetAPIIsUnhealthy generates a mocked unhealthy response for dataset API healthecheck
func (c *Component) datasetAPIIsUnhealthy() error {
	const res = `{"status": "CRITICAL"}`
	c.DatasetAPI.NewHandler().
		Get("/health").
		Reply(http.StatusInternalServerError).
		BodyString(res)
	return nil
}

// importAPIIsHealthy generates a mocked healthy response for import API healthecheck
func (c *Component) importAPIIsHealthy() error {
	const res = `{"status": "OK"}`
	c.ImportAPI.NewHandler().
		Get("/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// cantabularServerIsHealthy generates a mocked healthy response for cantabular server
func (c *Component) cantabularServerIsHealthy() error {
	const res = `{"status": "OK"}`
	c.CantabularSrv.NewHandler().
		Get("/v10/datasets").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// cantabularAPIExtIsHealthy generates a mocked healthy response for cantabular server
func (c *Component) cantabularAPIExtIsHealthy() error {
	const res = `{"status": "OK"}`
	c.CantabularApiExt.NewHandler().
		Get("/graphql?query={}").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
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

// theFollowingCantabularCategoriesAreAvailable generates a mocked response for Cantabular Server
// GET /v1/codebook/{name} with the provided query
func (c *Component) theFollowingCantabularCategoriesAreAvailable(dataset string, variable string, cb *godog.DocString) error {
	// Encode the graphQL query with the provided dataset and variables
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": cantabular.QueryDimensionOptions,
		"variables": map[string]interface{}{
			"dataset":   dataset,
			"category":  "",
			"limit":     20,
			"offset":    0,
			"text":      "",
			"variables": []string{variable},
		},
	}); err != nil {
		return fmt.Errorf("failed to encode GraphQL query: %w", err)
	}

	// create graphql handler with expected query body
	c.CantabularApiExt.NewHandler().
		Post("/graphql").
		AssertBody(b.Bytes()).
		Reply(http.StatusOK).
		BodyString(cb.Content)

	return nil
}

// theCallToAddInstanceDimensionIsSuccessful generates a mocked successful response for Dataset API
// POST /instances/{id}/dimensions
func (c *Component) theCallToAddInstanceDimensionIsSuccessful(id string) error {
	c.DatasetAPI.NewHandler().
		Patch("/instances/"+id+"/dimensions").
		Reply(http.StatusOK).
		AddHeader("ETag", testETag)

	return nil
}

// theCallToAddInstanceDimensionIsUnsuccessful generates a mocked unsuccessful response for Dataset API
// POST /instances/{id}/dimensions
func (c *Component) theCallToAddInstanceDimensionIsUnsuccessful(id string) error {
	c.DatasetAPI.NewHandler().
		Patch("/instances/" + id + "/dimensions").
		Reply(http.StatusInternalServerError)

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

// theseInstanceCompleteEventsAreProduced consumes kafka messages that are expected to be produced by the service under test
// and validates that they match the expected values in the test
func (c *Component) theseInstanceCompleteEventsAreProduced(events *godog.Table) error {
	expected, err := assistdog.NewDefault().CreateSlice(new(event.InstanceComplete), events)
	if err != nil {
		return fmt.Errorf("failed to create slice from godog table: %w", err)
	}

	var got []*event.InstanceComplete
	listen := true

	c.consumer.StateWait(kafka.Consuming)
	log.Info(c.ctx, "component test consumer is consuming")

	for listen {
		select {
		case <-time.After(WaitEventTimeout):
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

func (c *Component) noInstanceCompleteEventsShouldBeProduced() error {
	listen := true

	c.consumer.StateWait(kafka.Consuming)
	log.Info(c.ctx, "component test consumer is consuming")

	for listen {
		select {
		case <-time.After(WaitEventTimeout):
			listen = false
		case <-c.consumer.Channels().Closer:
			return errors.New("closer channel closed")
		case msg, ok := <-c.consumer.Channels().Upstream:
			if !ok {
				return errors.New("upstream channel closed")
			}

			err := fmt.Errorf("unexpected message receieved: %s", msg.GetData())

			msg.Commit()
			msg.Release()

			return err
		}
	}

	return nil
}

func (c *Component) thisCategoryDimensionImportEventIsQueued(input *godog.DocString) error {
	// testing kafka message that will be produced
	var testEvent event.CategoryDimensionImport
	if err := json.Unmarshal([]byte(input.Content), &testEvent); err != nil {
		return fmt.Errorf("error unmarshaling input to event: %w body: %s", err, input.Content)
	}

	log.Info(c.ctx, "event to marshal: ", log.Data{
		"event": testEvent,
	})

	// marshal and send message
	b, err := schema.CategoryDimensionImport.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event from schema: %w", err)
	}

	log.Info(c.ctx, "marshalled event: ", log.Data{
		"event": b,
	})

	c.producer.Channels().Output <- b

	return nil
}
