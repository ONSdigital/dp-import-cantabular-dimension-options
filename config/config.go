package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-import-cantabular-dimension-options
type Config struct {
	BindAddr                          string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout           time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval               time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout        time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	KafkaAddr                         []string      `envconfig:"KAFKA_ADDR"                     json:"-"`
	KafkaVersion                      string        `envconfig:"KAFKA_VERSION"`
	KafkaOffsetOldest                 bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	KafkaNumWorkers                   int           `envconfig:"KAFKA_NUM_WORKERS"`
	KafkaMaxBytes                     int           `envconfig:"KAFKA_MAX_BYTES"`
	KafkaCategoryDimensionImportGroup string        `envconfig:"KAFKA_CATEGORY_DIMENSION_IMPORT_GROUP"`
	KafkaCategoryDimensionImportTopic string        `envconfig:"KAFKA_CATEGORY_DIMENSION_IMPORT_TOPIC"`
	KafkaInstanceCompleteTopic        string        `envconfig:"KAFKA_INSTANCE_COMPLETE_TOPIC"`
	DatasetAPIURL                     string        `envconfig:"DATASET_API_URL"`
	ImportAPIURL                      string        `envconfig:"IMPORT_API_URL"`
	CantabularURL                     string        `envconfig:"CANTABULAR_URL"`
	ServiceAuthToken                  string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
	ComponentTestUseLogFile           bool          `envconfig:"COMPONENT_TEST_USE_LOG_FILE"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                          ":26200",
		GracefulShutdownTimeout:           5 * time.Second,
		HealthCheckInterval:               30 * time.Second,
		HealthCheckCriticalTimeout:        90 * time.Second,
		KafkaAddr:                         []string{"localhost:9092"},
		KafkaVersion:                      "1.0.2",
		KafkaOffsetOldest:                 true,
		KafkaNumWorkers:                   1,
		KafkaMaxBytes:                     2000000,
		KafkaCategoryDimensionImportGroup: "dp-import-cantabular-dimension-options",
		KafkaCategoryDimensionImportTopic: "cantabular-dataset-category-dimension-import",
		KafkaInstanceCompleteTopic:        "cantabular-dataset-instance-complete",
		DatasetAPIURL:                     "http://localhost:22000",
		CantabularURL:                     "http://localhost:8491",
		ImportAPIURL:                      "http://localhost:21800",
		ServiceAuthToken:                  "",
		ComponentTestUseLogFile:           true,
	}

	return cfg, envconfig.Process("", cfg)
}
