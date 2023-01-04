# Redis Infrastructure Monitoring Example

This example provides a `docker-compose` environment that sends redis data to stdout and sfx. You can change the exporters to your liking by modifying `otel-collector-config.yaml`.

You'll need to install docker at a minimum. Ensure the following environment variables are properly set:

1. `REDIS_PASSWORD` (default: `changeme`, see `redis.conf`)
1. `SPLUNK_ACCESS_TOKEN` (for sfx exporter)
1. `SPLUNK_REALM` (for sfx exporter)

Once you've verified your environment, you can run the example by

```bash
$> docker-compose up
```

