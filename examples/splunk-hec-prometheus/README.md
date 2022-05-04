# Scraping Prometheus metrics from a static endpoint

This example showcases (and tests) the collector scraping a specific, custom made Prometheus endpoint, served from a static web server.

The example runs as a Docker Compose deployment. The collector can be configured to send metrics to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [analytics workspace](http://localhost:18000/en-US/app/search/analytics_workspace) to see the metrics collected by Splunk.
