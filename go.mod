module github.com/ONSdigital/dp-import-cantabular-dimension-options

go 1.16

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.1.7-beta
	github.com/ONSdigital/dp-component-test v0.3.1
	github.com/ONSdigital/dp-healthcheck v1.0.5
	github.com/ONSdigital/dp-import v1.2.1
	github.com/ONSdigital/dp-kafka/v2 v2.3.0
	github.com/ONSdigital/dp-net v1.0.12
	github.com/ONSdigital/log.go/v2 v2.0.5
	github.com/cucumber/godog v0.11.0
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/maxcnunes/httpfake v1.2.3
	github.com/pkg/errors v0.9.1
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/smartystreets/goconvey v1.6.4
)
