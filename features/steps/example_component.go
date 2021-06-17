package steps

import (
	"context"
	"fmt"
	"net/http"
	"os"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service/mock"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	dphttp "github.com/ONSdigital/dp-net/http"
)

type Component struct {
	componenttest.ErrorFeature
	KafkaConsumer kafka.IConsumerGroup
	killChannel   chan os.Signal
	apiFeature    *componenttest.APIFeature
	errorChan     chan error
	svc           *service.Service
	cfg           *config.Config
}

func NewComponent() *Component {

	c := &Component{errorChan: make(chan error)}

	consumer := kafkatest.NewMessageConsumer(false)
	consumer.CheckerFunc = funcCheck
	c.KafkaConsumer = consumer

	cfg, err := config.Get()
	if err != nil {
		return nil
	}

	c.cfg = cfg

	service.GetKafkaConsumer = c.GetConsumer
	service.GetHealthCheck = c.GetHealthCheck
	service.GetHTTPServer = c.GetHTTPServer

	c.svc = service.New()
	err = c.svc.Init(context.Background(), cfg, "", "", "")
	if err != nil {
		panic(fmt.Errorf("unexpected service Init error in NewComponent: %w", err))
	}

	return c
}

func (c *Component) Close() {
	os.Remove(c.cfg.OutputFilePath)
}

func (c *Component) Reset() {
	os.Remove(c.cfg.OutputFilePath)
}

func (c *Component) GetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return &mock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *Component) GetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	return dphttp.NewServer(bindAddr, router)
}

func (c *Component) GetConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	return c.KafkaConsumer, nil
}

func funcCheck(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}
