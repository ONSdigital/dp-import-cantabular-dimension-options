package steps

import (
	"context"
	"fmt"
	"os"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	kafka "github.com/ONSdigital/dp-kafka/v2"
)

var (
	BuildTime string = "1625046891"
	GitCommit string = "7434fe334d9f51b7239f978094ea29d10ac33b16"
	Version   string = ""
)

type Component struct {
	componenttest.ErrorFeature
	KafkaConsumer kafka.IConsumerGroup
	apiFeature    *componenttest.APIFeature
	errorChan     chan error
	svc           *service.Service
	cfg           *config.Config
}

func NewComponent() *Component {
	c := &Component{errorChan: make(chan error)}

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return nil
	}
	c.cfg = cfg

	// Create service and initialise it
	c.svc = service.New()
	err = c.svc.Init(context.Background(), cfg, BuildTime, GitCommit, Version)
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
