# HEC Traces Example

This example showcases how the agent works with Splunk Enterprise and traces.

The example runs as a Docker Compose deployment. The collector can be configured to send traces to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up --build
```
You can stop the example by pressing Ctrl+C.

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [spans comparisons dashboard](http://localhost:18000/en-US/app/search/spans_time_comparisons) to see a comparison of traces over time sent by the OpenTelemetry Collector.

You can also search traces for information, such as querying by tag or any element of traces, by searching the traces index `index=traces`

![Traces dashboard](traces_comparison_dashboard.png)