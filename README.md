dp-import-cantabular-dimension-options
================
dp-import-cantabular-dimension-options

### Getting started

* Run `make debug`

The service runs in the background consuming messages from Kafka.
An example event can be created using the helper script, `make produce`.

### Dependencies

* Requires running…
  * [kafka](https://github.com/ONSdigital/dp/blob/main/guides/INSTALLING.md#prerequisites)
* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable            | Default                                      | Description
| ------------------------------- | -------------------------------------------- | -----------
| BIND_ADDR                       | localhost:26200                              | The host and port to bind to
| GRACEFUL_SHUTDOWN_TIMEOUT       | 5s                                           | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL            | 30s                                          | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT    | 90s                                          | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| KAFKA_ADDR                      | localhost:9092                               | The kafka broker addresses (can be comma separated)
| KAFKA_VERSION                   | "1.0.2"                                      | The kafka version that this service expects to connect to
| KAFKA_OFFSET_OLDEST             | true                                         | Start processing Kafka messages in order from the oldest in the queue
| KAFKA_NUM_WORKERS               | 1                                            | The maximum number of parallel kafka consumers
| CATEGORY_DIMENSION_IMPORT_GROUP | dp-import-cantabular-dimension-options       | The consumer group this application to consume ImageUploaded messages
| CATEGORY_DIMENSION_IMPORT_TOPIC | cantabular-dataset-category-dimension-import | The name of the topic to consume messages from
| INSTANCE_COMPLETE_TOPIC         | cantabular-dataset-instance-complete         | The name of the topic to produce
| KAFKA_SEC_PROTO                 | _unset_                                      | if set to `TLS`, kafka connections will use TLS [[1]](#notes_1)
| KAFKA_SEC_CA_CERTS              | _unset_                                      | CA cert chain for the server cert [[1]](#notes_1)
| KAFKA_SEC_CLIENT_KEY            | _unset_                                      | PEM for the client key [[1]](#notes_1)
| KAFKA_SEC_CLIENT_CERT           | _unset_                                      | PEM for the client certificate [[1]](#notes_1)
| KAFKA_SEC_SKIP_VERIFY           | false                                        | ignores server certificate issues if `true` [[1]](#notes_1)

**Notes:**

    1. <a name="notes_1">For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)</a>


### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:8125/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

