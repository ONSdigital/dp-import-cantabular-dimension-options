package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":26200")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.CantabularURL, ShouldEqual, "http://localhost:8491")
				So(cfg.ImportAPIURL, ShouldEqual, "http://localhost:21800")
				So(cfg.CantabularHealthcheckEnabled, ShouldBeTrue)
				So(cfg.ServiceAuthToken, ShouldEqual, "")
				So(cfg.ComponentTestUseLogFile, ShouldBeFalse)
				So(cfg.BatchSizeLimit, ShouldEqual, 100)
				So(cfg.StopConsumingOnUnhealthy, ShouldBeTrue)
				So(cfg.KafkaConfig.Addr, ShouldResemble, []string{"localhost:9092", "localhost:9093", "localhost:9094"})
				So(cfg.KafkaConfig.ConsumerMinBrokersHealthy, ShouldEqual, 1)
				So(cfg.KafkaConfig.ProducerMinBrokersHealthy, ShouldEqual, 2)
				So(cfg.KafkaConfig.Version, ShouldEqual, "1.0.2")
				So(cfg.KafkaConfig.OffsetOldest, ShouldBeTrue)
				So(cfg.KafkaConfig.NumWorkers, ShouldEqual, 1)
				So(cfg.KafkaConfig.MaxBytes, ShouldEqual, 2000000)
				So(cfg.KafkaConfig.SecProtocol, ShouldEqual, "")
				So(cfg.KafkaConfig.SecCACerts, ShouldEqual, "")
				So(cfg.KafkaConfig.SecClientKey, ShouldEqual, "")
				So(cfg.KafkaConfig.SecClientCert, ShouldEqual, "")
				So(cfg.KafkaConfig.SecSkipVerify, ShouldBeFalse)
				So(cfg.KafkaConfig.CategoryDimensionImportGroup, ShouldEqual, "dp-import-cantabular-dimension-options")
				So(cfg.KafkaConfig.CategoryDimensionImportTopic, ShouldEqual, "cantabular-dataset-category-dimension-import")
				So(cfg.KafkaConfig.InstanceCompleteTopic, ShouldEqual, "cantabular-dataset-instance-complete")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
