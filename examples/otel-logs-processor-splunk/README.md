# Splunk Index Routing Example

This example showcases how the collector can collect data from files and send it to Splunk Enterprise.

In particular, this example uses 3 separate files that log different content.

Based on a regular expression of the content, the log data is routed to different Splunk indexes.

We show both how to transform the data and add an attribute to the original log message, as well as processing data 
and adding an attribute to the record based on the log record's attributes.

The example runs as a Docker Compose deployment. The collector can be configured to send logs to Splunk Enterprise.

Splunk is configured to receive data from the OpenTelemetry Collector using the HTTP Event collector. To learn more about HEC, visit [our guide](https://dev.splunk.com/enterprise/docs/dataapps/httpeventcollector/).

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Splunk will become available on port 18000. You can login on [http://localhost:18000](http://localhost:18000) with `admin` and `changeme`.

Once logged in, visit the [search application](http://localhost:18000/en-US/app/search) to see the logs collected by Splunk.

You can visit the `logs`, `logs2` and `logs3` indexes to see different log messages.

Here is how to see the contents of an index. In the search application, click on the Search bar at the top.

In the search bar, enter `index="logs"` to search the contents of the logs index.
