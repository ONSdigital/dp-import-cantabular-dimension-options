package event

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// TODO: remove or replace hello called logic with app specific

// Consume converts messages to event instances, and pass the event to the provided handler.
func Consume(ctx context.Context, cg kafka.IConsumerGroup, h Handler, numWorkers int) {
	// consume loop, to be executed by each worker
	var consume = func(workerID int) {
		for {
			select {
			case msg, ok := <-cg.Channels().Upstream:
				if !ok {
					log.Info(ctx, "upstream channel closed - closing event consumer loop", log.Data{"worker_id": workerID})
					return
				}

				if err := processMessage(context.Background(), msg, h); err != nil {
					log.Error(ctx, "failed to process message", err, log.Data{
						"status_code": statusCode(err),
						"log_data":    unwrapLogData(err),
					})
				}

				msg.Release()
			case <-cg.Channels().Closer:
				log.Event(ctx, "closing event consumer loop because closer channel is closed", log.INFO, log.Data{"worker_id": workerID})
				return
			}
		}
	}

	// workers to consume messages in parallel
	for w := 1; w <= numWorkers; w++ {
		go consume(w)
	}
}

// processMessage unmarshals the provided kafka message into an event and calls the handler.
// After the message is handled, it is committed, by default even on error to prevent reconsumption
// of dead messages.
func processMessage(ctx context.Context, msg kafka.Message, h Handler) error {
	defer msg.Commit()

	var e HelloCalled
	s := schema.HelloCalledEvent

	// unmarshal - commit on failure (consuming the message again would result in the same error)
	if err := s.Unmarshal(msg.GetData(), &e); err != nil {
		return &Error{
			err: fmt.Errorf("failed to unmarshal event: %w", err),
			logData: log.Data{
				"msg_data": msg.GetData(),
			},
		}
	}

	log.Info(ctx, "event received", log.Data{"event": e})

	if err := h.Handle(ctx, &e); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	log.Info(ctx, "event processed - committing message", log.Data{"event": e})
	return nil
}
