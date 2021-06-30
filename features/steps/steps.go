package steps

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^these hello events are consumed:$`, c.theseHelloEventsAreConsumed)
	ctx.Step(`^I should receive a hello-world response$`, c.iShouldReceiveAHelloworldResponse)
}

func (c *Component) iShouldReceiveAHelloworldResponse() error {
	// TODO implement test validation here
	return c.StepError()
}

func (c *Component) theseHelloEventsAreConsumed(table *godog.Table) error {
	ctx := context.Background()
	signals := registerInterrupt()

	// run application in separate goroutine
	serviceWg := &sync.WaitGroup{}
	serviceWg.Add(1)
	go func() {
		defer serviceWg.Done()
		c.svc.Start(context.Background(), c.errorChan)

		// blocks until an os interrupt or a fatal error occurs
		select {
		case err := <-c.errorChan:
			err = fmt.Errorf("service error received: %w", err)
			c.svc.Close(ctx)
			panic(fmt.Errorf("unexpected error received from errorChan: %w", err))
		case sig := <-signals:
			log.Info(ctx, "os signal received", log.Data{"signal": sig})
		}
		if err := c.svc.Close(ctx); err != nil {
			panic(fmt.Errorf("unexpected error during service graceful shutdown: %w", err))
		}
	}()

	// testing kafka message that will be produced
	var testEvent = event.CategoryDimensionImport{
		JobID:       "testJobID",
		InstanceID:  "testInstanceID",
		DimensionID: "testDimensionID",
	}

	// producer
	p, err := kafka.NewProducer(
		ctx,
		c.cfg.KafkaAddr,
		c.cfg.CategoryDimensionImportTopic,
		kafka.CreateProducerChannels(),
		&kafka.ProducerConfig{
			KafkaVersion:    &c.cfg.KafkaVersion,
			MaxMessageBytes: &c.cfg.KafkaMaxBytes,
		},
	)
	if err != nil {
		panic(fmt.Errorf("unexpected error in NewComponent wile creating a kafka producer: %w", err))
	}
	p.Channels().LogErrors(ctx, "producer")

	// wait for producer to be ready
	<-p.Channels().Ready
	log.Info(ctx, "producer ready. Sending message")

	// send message
	validMessage := kafkatest.NewMessage(marshal(testEvent), 1)
	p.Channels().Output <- validMessage.GetData()

	// close producer
	if err := p.Close(ctx); err != nil {
		panic(fmt.Errorf("unexpected error in NewComponent wile closing a kafka producer: %w", err))
	}

	// give enough time to the app to process the message
	// TODO is there a way to communicate with channels and/or wait groups that doesn't affect too much the code? It would be better than sleeping
	time.Sleep(300 * time.Millisecond)

	// kill application
	signals <- os.Interrupt

	// wait for graceful shutdown to finish (or timeout)
	serviceWg.Wait()

	log.Info(ctx, "theseHelloEventsAreConsumed done")
	return nil
}

func registerInterrupt() chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals
}

// marshal helper method to marshal a event into a []byte
func marshal(event event.CategoryDimensionImport) []byte {
	bytes, err := schema.CategoryDimensionImport.Marshal(event)
	if err != nil {
		panic(fmt.Errorf("unexpected error during marshal: %w", err))
	}
	return bytes
}
