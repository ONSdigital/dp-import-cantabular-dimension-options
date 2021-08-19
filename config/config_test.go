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
				So(cfg.KafkaAddr, ShouldHaveLength, 1)
				So(cfg.KafkaMaxBytes, ShouldEqual, 2000000)
				So(cfg.KafkaAddr[0], ShouldEqual, "localhost:9092")
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaNumWorkers, ShouldEqual, 1)
				So(cfg.KafkaCategoryDimensionImportGroup, ShouldEqual, "dp-import-cantabular-dimension-options")
				So(cfg.KafkaCategoryDimensionImportTopic, ShouldEqual, "cantabular-dataset-category-dimension-import")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.CantabularURL, ShouldEqual, "http://localhost:8491")
				So(cfg.ImportAPIURL, ShouldEqual, "http://localhost:21800")
				So(cfg.CantabularHealthcheckEnabled, ShouldBeFalse)
				So(cfg.ServiceAuthToken, ShouldEqual, "")
				So(cfg.ComponentTestUseLogFile, ShouldBeFalse)
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
