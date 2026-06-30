# Timestamp Processor Deprecation

The Splunk `timestamp` processor is deprecated. Use the upstream OpenTelemetry Collector
[`transform` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor)
instead.

The `timestamp` processor applies a single configured duration offset to telemetry timestamps.
The `transform` processor can apply the same offset with OTTL statements and should be used for
new configurations.

## Migration

Replace a `timestamp` processor configuration such as:

```yaml
processors:
  timestamp/add2h:
    offset: 2h
```

with:

```yaml
processors:
  transform/timestamp_offset:
    error_mode: propagate
    trace_statements:
      - context: span
        statements:
          - set(span.start_time, span.start_time + Duration("2h")) where span.start_time_unix_nano != 0
          - set(span.end_time, span.end_time + Duration("2h")) where span.end_time_unix_nano != 0
      - context: spanevent
        statements:
          - set(spanevent.time, spanevent.time + Duration("2h")) where spanevent.time_unix_nano != 0

    log_statements:
      - context: log
        statements:
          - set(log.time, log.time + Duration("2h")) where log.time_unix_nano != 0
          - set(log.observed_time, log.observed_time + Duration("2h")) where log.observed_time_unix_nano != 0

    metric_statements:
      - context: datapoint
        statements:
          - set(datapoint.start_time, datapoint.start_time + Duration("2h")) where datapoint.start_time_unix_nano != 0
          - set(datapoint.time, datapoint.time + Duration("2h")) where datapoint.time_unix_nano != 0
      - context: exemplar
        statements:
          - set(exemplar.time, exemplar.time + Duration("2h")) where exemplar.time_unix_nano != 0
```

Then update pipelines to use `transform/timestamp_offset` wherever they used the `timestamp`
processor instance.

For a negative offset, use a negative duration:

```yaml
processors:
  transform/timestamp_offset:
    error_mode: propagate
    log_statements:
      - context: log
        statements:
          - set(log.time, log.time + Duration("-3h")) where log.time_unix_nano != 0
          - set(log.observed_time, log.observed_time + Duration("-3h")) where log.observed_time_unix_nano != 0
```

## Coverage

The migration statements preserve the `timestamp` processor behavior for:

- log record timestamps and observed timestamps
- span start and end timestamps
- span event timestamps
- metric datapoint start and sample timestamps
- metric exemplar timestamps

Keep the `where ... != 0` guards. The `timestamp` processor leaves unset timestamps unchanged,
and the guards preserve that behavior in `transform`.

If a pipeline only carries one signal type, include only the matching `trace_statements`,
`log_statements`, or `metric_statements` block.

## Timeline

- **Deprecation notice**: Current release
- **Planned removal**: Future release, to be announced
