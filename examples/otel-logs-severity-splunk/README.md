# Splunk Severity Mapping Example

This example showcases how the collector interprets the content of log lines and assigns them severity before sending to Splunk Enterprise.

In particular, this example uses 3 separate files that log different content, each with a different "Level" field.

We parse each line as a JSON object and extract the value of the Level field to indicate the severity of the record.

The example runs as a Docker Compose deployment. The collector can be configured to send logs to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [search application](http://localhost:18000/en-US/app/search) to see the logs collected by Splunk.

You can visit the `logs` indexes to see messages assigned different severity levels.

In the search application, click on the Search bar at the top.

In the search bar, enter `index="logs"` to search the contents of the logs index.
