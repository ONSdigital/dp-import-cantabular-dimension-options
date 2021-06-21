package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-import-cantabular-dimension-options"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Get Config
	config, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Create Kafka Consumer
	cgChannels := kafka.CreateConsumerGroupChannels(1)
	kafkaOffset := kafka.OffsetOldest
	kafkaConsumer, err := kafka.NewConsumerGroup(
		ctx,
		config.KafkaAddr,
		config.InstanceCompleteTopic,
		"test-consumer-group",
		cgChannels,
		&kafka.ConsumerGroupConfig{
			KafkaVersion: &config.KafkaVersion,
			Offset:       &kafkaOffset,
		},
	)
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka consumer", err, log.Data{"topic": config.InstanceCompleteTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaConsumer.Channels().LogErrors(ctx, "kafka consumer")

	// block until the consumer is initialised (of closed)
	select {
	case <-cgChannels.Ready:
		log.Warn(ctx, "[KAFKA-TEST] Consumer is now initialised.")
	case <-cgChannels.Closer:
		log.Warn(ctx, "[KAFKA-TEST] Consumer is being closed.")
	}

	// consume in a new loop
	go consume(ctx, cgChannels.Upstream)

	// blocks until an os interrupt or a fatal error occurs
	sig := <-signals
	log.Info(ctx, "os signal received", log.Data{"signal": sig})
	closeConsumerGroup(ctx, kafkaConsumer, config.GracefulShutdownTimeout)
}

// consume waits for messages to arrive to the upstream channel and consumes them, in an infinite loop
func consume(ctx context.Context, upstream chan kafka.Message) {
	log.Info(ctx, "worker started consuming")
	for {
		consumedMessage, ok := <-upstream
		if !ok {
			break
		}
		logData := log.Data{"messageOffset": consumedMessage.Offset()}
		log.Info(ctx, "[KAFKA-TEST] Received message", logData)

		// TODO when schema is defined, we can unmarshal and log the message here
		consumedData := consumedMessage.GetData()
		logData["messageString"] = string(consumedData)
		logData["messageRaw"] = consumedData
		logData["messageLen"] = len(consumedData)

		consumedMessage.CommitAndRelease()
		log.Info(ctx, "[KAFKA-TEST] committed and released message", logData)
	}
}

func closeConsumerGroup(ctx context.Context, cg *kafka.ConsumerGroup, gracefulShutdownTimeout time.Duration) error {
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": gracefulShutdownTimeout})
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)

	// track shutown gracefully closes up
	var hasShutdownError bool

	// background graceful shutdown
	go func() {
		defer cancel()
		log.Info(ctx, "[KAFKA-TEST] Closing kafka consumerGroup")
		if err := cg.Close(ctx); err != nil {
			hasShutdownError = true
		}
		log.Info(ctx, "[KAFKA-TEST] Closed kafka consumerGroup")
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "[KAFKA-TEST] graceful shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}
