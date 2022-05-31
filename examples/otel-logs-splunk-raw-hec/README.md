# Splunk Raw HEC example

This example showcases how the collector can expose a HTTP endpoint to receive data over the HEC endpoint as a raw payload.

To learn more about HEC raw capabilities, head on to the [HEC documentation](https://docs.splunk.com/Documentation/Splunk/8.2.6/Data/FormateventsforHTTPEventCollector#Raw_event_parsing).

The example runs as a Docker Compose deployment.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [search application](http://localhost:18000/en-US/app/search) to see the logs collected by Splunk.

# Using curl locally

You can send logs to Splunk by sending data via curl with the following command:

```bash
$> curl -XPOST -k localhost:18088/services/collector/raw -d "your message here"
```

* `18088` is the port we expose our collector's HEC endpoint on the host machine.
* `/services/collector/raw` is the path to the HEC raw parsing entrypoint.
* `-k` is a required flag as we talk to localhost, and therefere no SSL certificate checks should take place.
* The message is sent as the body of a POST request, so `-XPOST` is necessary, and the `-d` flag indicates the data to send.