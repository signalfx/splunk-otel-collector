job "otel-gateway" {
  datacenters = ["dc1"]
  type        = "service"

  constraint {
    attribute = "${attr.nomad.version}"
    operator  = "semver"
    value     = "< 1.3.0"
  }

  group "otel-gateway" {
    count = 1

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

      # Extensions
      port "health-check" {
        to = 13133
      }

      port "zpages" {
        to = 55679
      }
    }

    service {
      name = "otel-gateway"
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
      name = "otel-gateway"
      port = "otlp"
      tags = ["otlp"]
    }

    service {
      name = "otel-gateway"
      port = "jaeger-grpc"
      tags = ["jaeger-grpc"]
    }

    service {
      name = "otel-gateway"
      port = "jaeger-thrift-http"
      tags = ["jaeger-thrift-http"]
    }

    service {
      name = "otel-gateway"
      port = "zipkin"
      tags = ["zipkin"]
    }

    service {
      name = "otel-gateway"
      port = "signalfx"
      tags = ["signalfx"]
    }

    service {
      name = "otel-gateway"
      port = "metrics"
      tags = ["metrics"]
    }

    service {
      name = "otel-gateway"
      port = "zpages"
      tags = ["zpages"]
    }

    task "otel-gateway" {
      driver = "docker"

      config {
        image = "quay.io/signalfx/splunk-otel-collector:latest"
        force_pull = false
        entrypoint = [
          "/otelcol",
          "--config=local/config/otel-gateway-config.yaml",
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
  http_forwarder:
    egress:
      endpoint: https://api.${SPLUNK_REALM}.signalfx.com
  zpages: null
receivers:
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
  signalfx:
    access_token_passthrough: true
    endpoint: 0.0.0.0:9943
  zipkin:
    endpoint: 0.0.0.0:9411
  prometheus/collector:
    config:
      scrape_configs:
      - job_name: otel-gateway
        scrape_interval: 10s
        static_configs:
        - targets:
          - ${HOSTNAME}:8889
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
    override: false
    timeout: 10s
exporters:
  sapm:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    endpoint: https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    api_url: https://api.${SPLUNK_REALM}.signalfx.com
    ingest_url: https://ingest.${SPLUNK_REALM}.signalfx.com
  logging:
    verbosity: detailed
service:
  extensions:
  - health_check
  - http_forwarder
  - zpages
  pipelines:
    logs/signalfx-events:
      exporters:
      - signalfx
      - logging
      processors:
      - memory_limiter
      - batch
      receivers:
      - signalfx
    metrics:
      exporters:
      - signalfx
      - logging
      processors:
      - memory_limiter
      - batch
      receivers:
      - otlp
      - signalfx
    traces:
      exporters:
      - sapm
      - logging
      processors:
      - memory_limiter
      - batch
      receivers:
      - otlp
      - jaeger
      - zipkin
    metrics/collector:
      exporters:
      - signalfx
      - logging
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - prometheus/collector
EOF
        destination = "local/config/otel-gateway-config.yaml"
      }
    }
  }
}
