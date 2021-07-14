package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler"
	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the event handler service
type Service struct {
	Cfg              *config.Config
	Server           HTTPServer
	HealthCheck      HealthChecker
	Consumer         kafka.IConsumerGroup
	Producer         kafka.IProducer
	CantabularClient CantabularClient
	DatasetAPIClient DatasetAPIClient
	ImportAPIClient  ImportAPIClient
}

// GetKafkaConsumer returns a Kafka consumer with the provided config
var GetKafkaConsumer = func(ctx context.Context, cfg *config.Config) (dpkafka.IConsumerGroup, error) {
	cgChannels := dpkafka.CreateConsumerGroupChannels(1)
	kafkaOffset := dpkafka.OffsetNewest
	if cfg.KafkaOffsetOldest {
		kafkaOffset = dpkafka.OffsetOldest
	}
	return dpkafka.NewConsumerGroup(
		ctx,
		cfg.KafkaAddr,
		cfg.CategoryDimensionImportTopic,
		cfg.CategoryDimensionImportGroup,
		cgChannels,
		&dpkafka.ConsumerGroupConfig{
			KafkaVersion: &cfg.KafkaVersion,
			Offset:       &kafkaOffset,
		},
	)
}

// GetKafkaProducer returns a kafka producer with the provided config
var GetKafkaProducer = func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	return kafka.NewProducer(
		ctx,
		cfg.KafkaAddr,
		cfg.InstanceCompleteTopic,
		kafka.CreateProducerChannels(),
		&kafka.ProducerConfig{
			KafkaVersion:    &cfg.KafkaVersion,
			MaxMessageBytes: &cfg.KafkaMaxBytes,
		},
	)
}

// GetHTTPServer returns an http server
var GetHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// GetHealthCheck returns a healthcheck
var GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

var GetCantabularClient = func(cfg *config.Config) CantabularClient {
	return cantabular.NewClient(
		dphttp.NewClient(),
		cantabular.Config{
			Host: cfg.CantabularURL,
		},
	)
}

var GetDatasetAPIClient = func(cfg *config.Config) DatasetAPIClient {
	return dataset.NewAPIClient(cfg.DatasetAPIURL)
}

var GetImportAPIClient = func(cfg *config.Config) ImportAPIClient {
	return importapi.New(cfg.ImportAPIURL)
}

// New creates a new empty service
func New() *Service {
	return &Service{}
}

// Init initialises all the service dependencies, including healthcheck with checkers, api and middleware
func (svc *Service) Init(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string) error {
	var err error

	if cfg == nil {
		return errors.New("nil config passed to service init")
	}

	svc.Cfg = cfg

	// Get Kafka consumer
	if svc.Consumer, err = GetKafkaConsumer(ctx, cfg); err != nil {
		return fmt.Errorf("failed to initialise kafka consumer: %w", err)
	}

	// Get Kafka producer
	svc.Producer, err = GetKafkaProducer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialise kafka producer: %w", err)
	}

	// Get API clients
	svc.CantabularClient = GetCantabularClient(cfg)
	svc.DatasetAPIClient = GetDatasetAPIClient(cfg)
	svc.ImportAPIClient = GetImportAPIClient(cfg)

	// Get HealthCheck
	if svc.HealthCheck, err = GetHealthCheck(cfg, buildTime, gitCommit, version); err != nil {
		return fmt.Errorf("could not instantiate healthcheck: %w", err)
	}

	if err := svc.registerCheckers(); err != nil {
		return fmt.Errorf("unable to register checkers: %w", err)
	}

	// Get HTTP Server with collectionID checkHeader
	r := mux.NewRouter()
	r.StrictSlash(true).Path("/health").HandlerFunc(svc.HealthCheck.Handler)
	svc.Server = GetHTTPServer(cfg.BindAddr, r)

	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {
	log.Info(ctx, "starting service...")

	// Start kafka error logging
	svc.Consumer.Channels().LogErrors(ctx, "error received from kafka consumer, topic: "+svc.Cfg.CategoryDimensionImportTopic)
	svc.Producer.Channels().LogErrors(ctx, "error received from kafka producer, topic: "+svc.Cfg.InstanceCompleteTopic)

	// Start consuming Kafka messages with the Event Handler
	event.Consume(
		ctx,
		svc.Consumer,
		handler.NewCategoryDimensionImport(
			*svc.Cfg,
			svc.CantabularClient,
			svc.DatasetAPIClient,
			svc.ImportAPIClient,
			svc.Producer,
		),
		svc.Cfg.KafkaNumWorkers,
	)

	// Start health checker
	svc.HealthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Cfg.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.HealthCheck != nil {
			svc.HealthCheck.Stop()
			log.Info(ctx, "stopped health checker")
		}

		// If kafka consumer exists, stop listening to it.
		// This will automatically stop the event consumer loops and no more messages will be processed.
		// The kafka consumer will be closed after the service shuts down.
		if svc.Consumer != nil {
			if err := svc.Consumer.StopListeningToConsumer(ctx); err != nil {
				log.Error(ctx, "error stopping kafka consumer listener", err)
				hasShutdownError = true
			}
			log.Info(ctx, "stopped kafka consumer listener")
		}

		// stop any incoming requests before closing any outbound connections
		if svc.Server != nil {
			if err := svc.Server.Shutdown(ctx); err != nil {
				log.Error(ctx, "failed to shutdown http server", err)
				hasShutdownError = true
			}
			log.Info(ctx, "stopped http server")
		}

		// If kafka producer exists, close it.
		if svc.Producer != nil {
			if err := svc.Producer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka producer", err)
				hasShutdownError = true
			}
			log.Event(ctx, "closed kafka producer", log.INFO)
		}

		// If kafka consumer exists, close it.
		if svc.Consumer != nil {
			if err := svc.Consumer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka consumer", err)
				hasShutdownError = true
			}
			log.Info(ctx, "closed kafka consumer")
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}

	// other error
	if hasShutdownError {
		return fmt.Errorf("failed to shutdown gracefully")
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

// registerCheckers adds the checkers for the service clients to the health check object.
func (svc *Service) registerCheckers() error {
	hc := svc.HealthCheck

	if err := hc.AddCheck("Kafka consumer", svc.Consumer.Checker); err != nil {
		return fmt.Errorf("error adding check for Kafka consumer: %w", err)
	}

	if err := hc.AddCheck("Kafka producer", svc.Producer.Checker); err != nil {
		return fmt.Errorf("error adding check for Kafka producer: %w", err)
	}

	if err := hc.AddCheck("Cantabular", svc.CantabularClient.Checker); err != nil {
		return fmt.Errorf("error adding check for Cantabular: %w", err)
	}

	if err := hc.AddCheck("Dataset API", svc.DatasetAPIClient.Checker); err != nil {
		return fmt.Errorf("error adding check for Dataset API: %w", err)
	}

	if err := hc.AddCheck("Import API", svc.ImportAPIClient.Checker); err != nil {
		return fmt.Errorf("error adding check for Import API: %w", err)
	}

	return nil
}
