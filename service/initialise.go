package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	dphttp "github.com/ONSdigital/dp-net/http"
)

// GetKafkaConsumer returns a Kafka consumer with the provided config
var GetKafkaConsumer = func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	kafkaOffset := kafka.OffsetNewest
	if cfg.KafkaConfig.OffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:       cfg.KafkaConfig.Addr,
		Topic:             cfg.KafkaConfig.CategoryDimensionImportTopic,
		GroupName:         cfg.KafkaConfig.CategoryDimensionImportGroup,
		MinBrokersHealthy: &cfg.KafkaConfig.ConsumerMinBrokersHealthy,
		KafkaVersion:      &cfg.KafkaConfig.Version,
		NumWorkers:        &cfg.KafkaConfig.NumWorkers,
		Offset:            &kafkaOffset,
	}
	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		cgConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}
	return kafka.NewConsumerGroup(ctx, cgConfig)
}

// GetKafkaProducer returns a kafka producer with the provided config
var GetKafkaProducer = func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       cfg.KafkaConfig.Addr,
		Topic:             cfg.KafkaConfig.InstanceCompleteTopic,
		MinBrokersHealthy: &cfg.KafkaConfig.ProducerMinBrokersHealthy,
		KafkaVersion:      &cfg.KafkaConfig.Version,
		MaxMessageBytes:   &cfg.KafkaConfig.MaxBytes,
	}
	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}
	return kafka.NewProducer(ctx, pConfig)
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
		cantabular.Config{
			Host: cfg.CantabularURL,
		},
		dphttp.NewClient(),
		nil,
	)
}

var GetDatasetAPIClient = func(cfg *config.Config) DatasetAPIClient {
	return dataset.NewAPIClient(cfg.DatasetAPIURL)
}

var GetImportAPIClient = func(cfg *config.Config) ImportAPIClient {
	return importapi.New(cfg.ImportAPIURL)
}
