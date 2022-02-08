# Tomcat and MySQL example

This example shows how the OpenTelemetry Collector can collect data from Apache Tomcat and MySQL, and send it to both Splunk Enterprise and Splunk APM.

## Set up

This example will download the sample.war file from the [Apache Tomcat website](https://tomcat.apache.org/tomcat-7.0-doc/appdev/sample/).


To deploy the example:
1. Check out the [Splunk OpenTelemetry Collector repository](https://github.com/signalfx/splunk-otel-collector).
2. Open a terminal.
3. Type the following commands:
```bash
$> cd examples/tomcat-mysql
$> curl https://tomcat.apache.org/tomcat-7.0-doc/appdev/sample/sample.war
$> docker-compose up
```
You can stop the example by pressing Ctrl + C.

Splunk Enterprise becomes available on port 18000. Log in to [http://localhost:18000](http://localhost:18000) with the user name `admin` and password `changeme`.

From there, you can see logs flowing in by searching for `index="logs"`.

You can visit `http://localhost:8080/sample` to visit the sample application. This will generate Apache Tomcat access logs.

