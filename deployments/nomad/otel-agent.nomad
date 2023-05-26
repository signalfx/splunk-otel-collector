job "otel-agent" {
  datacenters = ["dc1"]
  type        = "system"

  constraint {
    attribute = "${attr.nomad.version}"
    operator  = "semver"
    value     = "< 1.3.0"
  }

  group "otel-agent" {
    network {
      port "metrics" {
        to = 8889
      }

      # Receivers
      port "otlp" {
        to = 4317
      }

      port "jaeger-grpc" {
        to = 14250
      }

      port "jaeger-thrift-http" {
        to = 14268
      }

      port "zipkin" {
        to = 9411
      }

      port "signalfx" {
        to = 9943
      }

      port "sfx-forwarder" {
        to = 9080
      }

      # Extensions
      port "health-check" {
        to = 13133
      }

      port "zpages" {
        to = 55679
      }
    }

    service {
      name = "otel-agent"
      port = "health-check"
      tags = ["health"]

      check {
        type     = "http"
        port     = "health-check"
        path     = "/"
        interval = "5s"
        timeout  = "2s"
      }
    }

    service {
      name = "otel-agent"
      port = "otlp"
      tags = ["otlp"]
    }

    service {
      name = "otel-agent"
      port = "jaeger-grpc"
      tags = ["jaeger-grpc"]
    }

    service {
      name = "otel-agent"
      port = "jaeger-thrift-http"
      tags = ["jaeger-thrift-http"]
    }

    service {
      name = "otel-agent"
      port = "zipkin"
      tags = ["zipkin"]
    }

    service {
      name = "otel-agent"
      port = "signalfx"
      tags = ["signalfx"]
    }

    service {
      name = "otel-agent"
      port = "metrics"
      tags = ["metrics"]
    }

    service {
      name = "otel-agent"
      port = "zpages"
      tags = ["zpages"]
    }

    service {
      name = "otel-agent"
      port = "sfx-forwarder"
      tags = ["sfx-forwarder"]
    }

    task "otel-agent" {
      driver = "docker"

      config {
        image = "quay.io/signalfx/splunk-otel-collector:latest"
        force_pull = false
        entrypoint = [
          "/otelcol",
          "--config=local/config/otel-agent-config.yaml",
          "--metrics-addr=0.0.0.0:8889",
        ]

        ports = [
          "metrics",
          "otlp",
          "jaeger-grpc",
          "jaeger-thrift-http",
          "zipkin",
          "health-check",
          "zpages",
          "signalfx",
          "sfx-forwarder",
        ]
      }

      env {
        SPLUNK_ACCESS_TOKEN = "<SPLUNK_ACCESS_TOKEN>"
        SPLUNK_REALM = "<SPLUNK_REALM>"
        SPLUNK_MEMORY_TOTAL_MIB = 500
      }

      resources {
        cpu    = 500
        memory = 500
      }

      template {
        data        = <<EOF
extensions:
  health_check: null
  zpages: null
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu: null
      disk: null
      filesystem: null
      load: null
      memory: null
      network: null
      paging: null
      processes: null
  jaeger:
    protocols:
      grpc:
        endpoint: 0.0.0.0:14250
      thrift_http:
        endpoint: 0.0.0.0:14268
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
  prometheus/agent:
    config:
      scrape_configs:
      - job_name: otel-agent
        scrape_interval: 10s
        static_configs:
        - targets:
          - ${HOSTNAME}:8889
  signalfx:
    endpoint: 0.0.0.0:9943
  smartagent/signalfx-forwarder:
    listenAddress: 0.0.0.0:9080
    type: signalfx-forwarder
  zipkin:
    endpoint: 0.0.0.0:9411
processors:
  batch: null
  memory_limiter:
    ballast_size_mib: ${SPLUNK_BALLAST_SIZE_MIB}
    check_interval: 2s
    limit_mib: ${SPLUNK_MEMORY_LIMIT_MIB}
  resourcedetection:
    detectors:
    - system
    - env
    override: true
    timeout: 10s
exporters:
  sapm:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    endpoint: https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    api_url: https://api.${SPLUNK_REALM}.signalfx.com
    correlation: null
    ingest_url: https://ingest.${SPLUNK_REALM}.signalfx.com
    sync_host_metadata: true
  logging:
    verbosity: detailed
service:
  extensions:
  - health_check
  - zpages
  pipelines:
    metrics:
      exporters:
      - logging
      - signalfx
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - hostmetrics
      - signalfx
    metrics/agent:
      exporters:
      - logging
      - signalfx
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - prometheus/agent
    traces:
      exporters:
      - logging
      - sapm
      - signalfx
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - otlp
      - jaeger
      - smartagent/signalfx-forwarder
      - zipkin
EOF
        destination = "local/config/otel-agent-config.yaml"
      }
    }
  }
}
