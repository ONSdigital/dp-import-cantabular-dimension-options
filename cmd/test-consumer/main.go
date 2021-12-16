package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	serviceName       = "kafka-example-consumer"
	consumerGroupName = "test-consumer-group"
)

var consumeCount = 0

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(ctx, "fatal runtime error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Get Config
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %s", err)
	}

	// run kafka Consumer Group
	kafkaConsumer, err := runConsumerGroup(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error running kafka consumer: %s", err)
	}

	// blocks until an os interrupt or a fatal error occurs
	sig := <-signals
	log.Info(ctx, "os signal received", log.Data{"signal": sig})
	return closeConsumerGroup(ctx, kafkaConsumer, cfg.GracefulShutdownTimeout)
}

// handle un-marshals and logs events received by the kafka consumer
func handle(ctx context.Context, workerID int, msg kafka.Message) error {
	consumeCount++
	msgData := msg.GetData()

	var e event.InstanceComplete
	if err := schema.InstanceComplete.Unmarshal(msgData, &e); err != nil {
		log.Error(ctx, "failed to unmarshal event", err)
	}

	log.Info(ctx, "Received message", log.Data{
		"event":  e,
		"offest": msg.Offset(),
		"len":    len(msgData),
		"count":  consumeCount,
	})
	return nil
}

func runConsumerGroup(ctx context.Context, cfg *config.Config) (*kafka.ConsumerGroup, error) {
	log.Info(ctx, "Starting ConsumerGroup (messages sent to stdout)", log.Data{"config": cfg})
	kafka.SetMaxMessageSize(int32(cfg.KafkaConfig.MaxBytes))

	// Create ConsumerGroup with channels and config
	kafkaOffset := kafka.OffsetOldest
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:  cfg.KafkaConfig.Addr,
		Topic:        cfg.KafkaConfig.InstanceCompleteTopic,
		GroupName:    consumerGroupName,
		KafkaVersion: &cfg.KafkaConfig.Version,
		Offset:       &kafkaOffset,
	}
	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		cgConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}
	cg, err := kafka.NewConsumerGroup(ctx, cgConfig)
	if err != nil {
		return nil, err
	}

	// register handler to log events
	cg.RegisterHandler(ctx, handle)

	// start consuming as soon as possible
	cg.Start()

	// go-routine to log errors from error channel
	cg.LogErrors(ctx)

	// Wait until the consumer is initialised (will return if it is already initialised)
	waitForInitialised(ctx, cg.Channels())

	return cg, nil
}

func closeConsumerGroup(ctx context.Context, cg *kafka.ConsumerGroup, gracefulShutdownTimeout time.Duration) error {
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": gracefulShutdownTimeout})
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)

	// track shutown gracefully closes up
	var hasShutdownError bool

	// background graceful shutdown
	go func() {
		defer cancel()
		log.Info(ctx, "Closing kafka consumerGroup")
		if err := cg.Close(ctx); err != nil {
			hasShutdownError = true
		}
		log.Info(ctx, "Closed kafka consumerGroup")
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "graceful shutdown timed out", ctx.Err())
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

// waitForInitialised blocks until the consumer is initialised or closed
func waitForInitialised(ctx context.Context, cgChannels *kafka.ConsumerGroupChannels) {
	select {
	case <-cgChannels.Initialised:
		log.Warn(ctx, "Consumer is now initialised.")
	case <-cgChannels.Closer:
		log.Warn(ctx, "Consumer is being closed.")
	}
}
