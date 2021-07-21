package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
)

const testETag = "13c7791bafdbaaf5e6660754feb1a58cd6aaa892"

// DimensionCompleteEventsTimeout is the maximum time that the testing producer will wait
// for DimensionComplete events to be produced before failing the test
var DimensionCompleteEventsTimeout = time.Second

// RegisterSteps maps the human-readable regular expressions to their corresponding funcs
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following instance with id "([^"]*)" is available from dp-dataset-api:$`, c.theFollowingInstanceIsAvailable)
	ctx.Step(`^the instance with id "([^"]*)" has ([^"]*) dimension options`, c.theInstanceHasDimensionCount)
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

// theInstanceHasDimensionCount generates a mocked response for dataset API
// GET /instances/{id}/dimensions with the provided totalCount and limit=offset=0
func (c *Component) theInstanceHasDimensionCount(id string, totalCount string) error {

	// count must be a number
	cnt, err := strconv.Atoi(totalCount)
	if err != nil {
		return nil
	}

	// response with the provided total count
	resp := fmt.Sprintf(`{
		"items": [],
		"count": 0,
		"offset": 0,
		"limit": 0,
		"total_count": %d
	}`, cnt)

	// create handler with the mocked response
	c.DatasetAPI.NewHandler().
		Get("/instances/"+id+"/dimensions").
		Reply(http.StatusOK).
		BodyString(resp).
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

	if !reflect.DeepEqual(got, expected) {
		return fmt.Errorf("\ngot:\n%s\n expected:\n%s",
			printValues(got),
			printValues(expected.([]*event.InstanceComplete)))
	}

	return nil
}

func printValues(in []*event.InstanceComplete) string {
	ret := "[ "
	for _, val := range in {
		ret += fmt.Sprintf("%+v ", val)
	}
	return ret + "]"
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
