# SignalFx Gateway Prometheus Remote write Receiver Example

This example provides a `docker-compose` environment that continually sends some fake prometheus remote writes to an otel receiver replacement for the deprecated SignalFx Gateway for Prometheus Remote Writes.
To run this, ensure you have `docker-compose` installed. 

## Configuration
You can change the exporters to your liking by modifying `otel-collector-config.yaml`.

Ensure the following environment variables are properly set, should you wish to send data to splunk observability cloud:
1. `SPLUNK_ACCESS_TOKEN`
2. `SPLUNK_REALM`

Alternatively, you can remove the `signalfx` array item from the `exporters` configuration map in `otel-collector-config.yaml`

Feel free to modify the sample client to your liking, or even disable it and write your own!

## Running
Once you've verified your environment, you can run the example by

```bash
$> docker-compose up
```

If everything is configured properly, logs with sample writes should start appearing in stdout shortly.
