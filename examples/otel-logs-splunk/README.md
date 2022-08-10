# Splunk HEC Example

This example showcases how the collector can collect data from files and send it to Splunk Enterprise.

The example runs as a Docker Compose deployment. The collector can be configured to send logs to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [search application](http://localhost:18000/en-US/app/search) to see the logs collected by Splunk.

If you would like to see the metrics exposed by the OpenTelemetry Collector, 
you can visit [http://localhost:8888](http://localhost:8888) to see the metrics emitted by the collector.

Here is a sample excerpt of metrics:
```
# HELP otelcol_exporter_enqueue_failed_log_records Number of log records failed to be added to the sending queue.
# TYPE otelcol_exporter_enqueue_failed_log_records counter
otelcol_exporter_enqueue_failed_log_records{exporter="splunk_hec/logs",service_instance_id="a266e329-b02f-459f-9ec7-2e06021d4135",service_version="v0.57.0"} 0
# HELP otelcol_exporter_enqueue_failed_metric_points Number of metric points failed to be added to the sending queue.
# TYPE otelcol_exporter_enqueue_failed_metric_points counter
otelcol_exporter_enqueue_failed_metric_points{exporter="splunk_hec/logs",service_instance_id="a266e329-b02f-459f-9ec7-2e06021d4135",service_version="v0.57.0"} 0
# HELP otelcol_exporter_enqueue_failed_spans Number of spans failed to be added to the sending queue.
# TYPE otelcol_exporter_enqueue_failed_spans counter
otelcol_exporter_enqueue_failed_spans{exporter="splunk_hec/logs",service_instance_id="a266e329-b02f-459f-9ec7-2e06021d4135",service_version="v0.57.0"} 0
# HELP otelcol_exporter_queue_capacity Fixed capacity of the retry queue (in batches)
# TYPE otelcol_exporter_queue_capacity gauge
otelcol_exporter_queue_capacity{exporter="splunk_hec/logs",service_instance_id="a266e329-b02f-459f-9ec7-2e06021d4135",service_version="v0.57.0"} 5000
# HELP otelcol_exporter_queue_size Current size of the retry queue (in batches)
# TYPE otelcol_exporter_queue_size gauge
otelcol_exporter_queue_size{exporter="splunk_hec/logs",service_instance_id="a266e329-b02f-459f-9ec7-2e06021d4135",service_version="v0.57.0"} 0
# HELP otelcol_exporter_sent_log_records Number of log record successfully sent to destination.
# TYPE otelcol_exporter_sent_log_records counter
otelcol_exporter_sent_log_records{exporter="splunk_hec/logs",service_instance_id="a266e329-b02f-459f-9ec7-2e06021d4135",service_version="v0.57.0"} 45
```

Additionally, this example runs Prometheus on port 9090. Visit [http://localhost:9090](http://localhost:9090), 
and enter a query to examine metrics exposed by the OpenTelemetry Collector.

For more information on the metrics exposed by the collector, read on to the [official OpenTelemetry monitoring documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/monitoring.md).

For more information on troubleshooting the collector, please refer to the [official OpenTelemetry documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/troubleshooting.md).