# Discovery confmap.Provider (Experimental)

**This component should not be considered stable. At this time its functionality is provided for testing and validation purposes only.**

The Discovery [confmap.Provider](https://pkg.go.dev/go.opentelemetry.io/collector/confmap#readme-provider) provides
the ability to define Collector service config through individual component yaml mappings in a `config.d` directory:

```mermaid
graph LR
  config.d[/config.d/] --> 1>service.yaml]
  subgraph 1a[service.yaml]
    1 --> 1a1[[pipelines:<br>metrics:<br>receivers:<br>- otlp<br>exporters:<br>- logging]]
  end
  config.d --> 2[/exporters/]
  subgraph 2a[exporters]
    2 --> 2a1>otlp.yaml]
    2a1 --> 2b1[[otlp:<br>endpoint: 1.2.3.4:2345]]
    2 --> 2a2>logging.yaml]
    2a2 --> 2b2[[logging:<br>logLevel: debug]]
  end
  config.d --> 3[/extensions/]
  subgraph 3a[extensions]
    3 --> 3a1>zpages.yaml]
    3a1 --> 3b1[[zpages:<br>endpoint: 0.0.0.0:12345]]
    3 --> 3a2>health-check.yaml]
    3a2 --> 3b2[[health_check:<br>path: /health]]
  end
  config.d --> 4[/processors/]
  subgraph 4a[processors]
    4 --> 4a1>batch.yaml]
    4a1 --> 4b1[[batch:<br>]]
    4 --> 4a2>resource-detection.yaml]
    4a2 --> 4b2[[resourcedetection:<br>detectors:<br>- system]]
  end
  config.d --> 5[/receivers/]
  subgraph 5a[receivers]
    5 --> 5a1>otlp.yaml]
    5a1 --> 5b1[[otlp:<br>protocols:<br>grpc]]
  end
```

This component is currently exposed in the Collector via the `--configd` option with corresponding
`--config-dir <config.d path>` and `SPLUNK_CONFIG_DIR` option and environment variable to load
additional components and service configuration from the specified `config.d` directory (`/etc/otel/collector/config.d`
by default).
