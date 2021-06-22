package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-import-cantabular-dimension-options"

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Error(ctx, "fatal runtime error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	svcErrors := make(chan error, 1)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("unable to retrieve configuration, error: %w", err)
	}

	// Run the service
	svc := service.New()
	if err := svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return fmt.Errorf("running service failed with error: %w", err)
	}
	svc.Start(ctx, svcErrors)

	// blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		err = fmt.Errorf("service error received: %w", err)
		svc.Close(ctx)
		return err
	case sig := <-signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}
	return svc.Close(ctx)
}
