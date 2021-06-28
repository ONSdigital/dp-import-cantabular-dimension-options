package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-import-cantabular-dimension-options
type Config struct {
	BindAddr                     string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout      time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval          time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout   time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	KafkaAddr                    []string      `envconfig:"KAFKA_ADDR"                     json:"-"`
	KafkaVersion                 string        `envconfig:"KAFKA_VERSION"`
	KafkaOffsetOldest            bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	KafkaNumWorkers              int           `envconfig:"KAFKA_NUM_WORKERS"`
	KafkaMaxBytes                int           `envconfig:"KAFKA_MAX_BYTES"`
	CategoryDimensionImportGroup string        `envconfig:"CATEGORY_DIMENSION_IMPORT_GROUP"`
	CategoryDimensionImportTopic string        `envconfig:"CATEGORY_DIMENSION_IMPORT_TOPIC"`
	InstanceCompleteTopic        string        `envconfig:"INSTANCE_COMPLETE_TOPIC"`
	OutputFilePath               string        `envconfig:"OUTPUT_FILE_PATH"` // TODO remove this when the handle logic is implemented
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                     "localhost:26200",
		GracefulShutdownTimeout:      5 * time.Second,
		HealthCheckInterval:          30 * time.Second,
		HealthCheckCriticalTimeout:   90 * time.Second,
		KafkaAddr:                    []string{"localhost:9092"},
		KafkaVersion:                 "1.0.2",
		KafkaOffsetOldest:            true,
		KafkaNumWorkers:              1,
		KafkaMaxBytes:                2000000,
		CategoryDimensionImportGroup: "dp-import-cantabular-dimension-options",
		CategoryDimensionImportTopic: "cantabular-dataset-category-dimension-import",
		InstanceCompleteTopic:        "cantabular-dataset-instance-complete",
		OutputFilePath:               "/tmp/helloworld.txt", // TODO remove this when the handle logic is implemented
	}

	return cfg, envconfig.Process("", cfg)
}
