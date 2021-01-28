# Prometheus Federation Endpoint Example

This example showcases how the agent works with Splunk Enterprise and an existing Prometheus deployment.

The example runs as a Docker Compose deployment. The collector can be configured to send various metrics to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up --build
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [analytics workspace](http://localhost:18000/en-US/app/search/analytics_workspace) to see which metrics are sent by the OpenTelemetry Collector.

Additionally, you can consult the [Prometheus UI](http://localhost:9090) to see the metric data collected from the sample go program.

# Diagram of the deployment

![Diagram](diagram.png)
