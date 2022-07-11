# OpenTelemetry log enrichment

This example showcases how the collector can add more metadata to log data being sent to Splunk.

The example runs as a Docker Compose deployment. The collector is configured to send logs to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [search application](http://localhost:18000/en-US/app/search) to see the logs collected by Splunk.

Enter the search `index=main` to search the logs ingested by Splunk.

You will see that logs have been tagged with additional information, a region and zone fields, set respectively to "normandy" and "eu-nor1".

The fields are set with the [attributes processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) as part of the Splunk OpenTelemetry Collector pipeline.