package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	serviceMock "github.com/ONSdigital/dp-import-cantabular-dimension-options/service/mock"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errKafkaConsumer = errors.New("Kafka consumer error")
	errKafkaProducer = fmt.Errorf("kafka producer error")
	errHealthcheck   = errors.New("healthCheck error")
	errServer        = fmt.Errorf("HTTP Server error")
	errAddCheck      = fmt.Errorf("healthcheck add check error")
)

func TestInit(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		consumerMock := &kafkatest.IConsumerGroupMock{}
		GetKafkaConsumer = func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
			return consumerMock, nil
		}

		producerMock := &kafkatest.IProducerMock{}
		GetKafkaProducer = func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
			return producerMock, nil
		}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		}
		GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
			return hcMock, nil
		}

		serverMock := &serviceMock.HTTPServerMock{}
		GetHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
			return serverMock
		}

		cantabularMock := &serviceMock.CantabularClientMock{
			CheckerFunc: func(context.Context, *healthcheck.CheckState) error {
				return nil
			},
		}
		GetCantabularClient = func(cfg *config.Config) CantabularClient { return cantabularMock }

		datasetAPIMock := &serviceMock.DatasetAPIClientMock{
			CheckerFunc: func(context.Context, *healthcheck.CheckState) error {
				return nil
			},
		}
		GetDatasetAPIClient = func(cfg *config.Config) DatasetAPIClient { return datasetAPIMock }

		svc := &Service{}

		Convey("Given that initialising Kafka consumer returns an error", func() {
			GetKafkaConsumer = func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
				return nil, errKafkaConsumer
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errKafkaConsumer)
				So(svc.cfg, ShouldResemble, cfg)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that initialising Kafka producer returns an error", func() {
			GetKafkaProducer = func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
				return nil, errKafkaProducer
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errKafkaProducer)
				So(svc.cfg, ShouldResemble, cfg)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
				return nil, errHealthcheck
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errHealthcheck)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.consumer, ShouldResemble, consumerMock)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			hcMock.AddCheckFunc = func(name string, checker healthcheck.Checker) error { return errAddCheck }

			Convey("Then service Init fails with the expected error", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errAddCheck), ShouldBeTrue)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.consumer, ShouldResemble, consumerMock)

				Convey("And other checkers don't try to register", func() {
					So(hcMock.AddCheckCalls(), ShouldHaveLength, 1)
				})
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			Convey("Then service Init succeeds, all dependencies are initialised", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.cfg, ShouldResemble, cfg)
				So(svc.consumer, ShouldResemble, consumerMock)
				So(svc.server, ShouldResemble, serverMock)

				Convey("And all checks are registered", func() {
					So(hcMock.AddCheckCalls(), ShouldHaveLength, 4)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Kafka consumer")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka producer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Cantabular")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Dataset API")
				})
			})
		})
	})
}

func TestStart(t *testing.T) {

	Convey("Having a correctly initialised Service with mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		consumerMock := &kafkatest.IConsumerGroupMock{
			ChannelsFunc: func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
		}

		producerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels { return &kafka.ProducerChannels{} },
		}

		hcMock := &serviceMock.HealthCheckerMock{
			StartFunc: func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{}

		svc := &Service{
			cfg:         cfg,
			server:      serverMock,
			healthCheck: hcMock,
			consumer:    consumerMock,
			producer:    producerMock,
		}

		Convey("When a service with a successful HTTP server is started", func() {
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return nil
			}
			serverWg.Add(1)
			svc.Start(ctx, make(chan error, 1))

			Convey("Then healthcheck is started and HTTP server starts listening", func() {
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a service with a failing HTTP server is started", func() {
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return errServer
			}
			errChan := make(chan error, 1)
			serverWg.Add(1)
			svc.Start(ctx, errChan)

			Convey("Then HTTP server errors are reported to the provided errors channel", func() {
				rxErr := <-errChan
				So(rxErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		consumerListening := true

		// kafka consumer mock, sets consumerListening when StopListeningToConsumer is called
		consumerMock := &kafkatest.IConsumerGroupMock{
			StopListeningToConsumerFunc: func(ctx context.Context) error {
				consumerListening = false
				return nil
			},
			CloseFunc: func(ctx context.Context) error { return nil },
		}

		// kafka producer mock, close will fail if consumer is still listening
		producerMock := &kafkatest.IProducerMock{
			CloseFunc: func(ctx context.Context) error {
				if consumerListening {
					return fmt.Errorf("Kafka producer closed while consumer is still listening")
				}
				return nil
			},
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			StopFunc: func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return fmt.Errorf("Server stopped before healthcheck")
				}
				return nil
			},
		}

		svc := &Service{
			cfg:         cfg,
			server:      serverMock,
			healthCheck: hcMock,
			consumer:    consumerMock,
			producer:    producerMock,
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(consumerMock.StopListeningToConsumerCalls(), ShouldHaveLength, 1)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(consumerMock.CloseCalls(), ShouldHaveLength, 1)
			So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(producerMock.CloseCalls(), ShouldHaveLength, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			consumerMock.StopListeningToConsumerFunc = func(ctx context.Context) error {
				return errKafkaConsumer
			}
			consumerMock.CloseFunc = func(ctx context.Context) error {
				return errKafkaConsumer
			}
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				return errServer
			}
			producerMock.CloseFunc = func(ctx context.Context) error {
				return errKafkaProducer
			}

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(consumerMock.StopListeningToConsumerCalls(), ShouldHaveLength, 1)
			So(consumerMock.CloseCalls(), ShouldHaveLength, 1)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(producerMock.CloseCalls(), ShouldHaveLength, 1)
		})
	})
}
