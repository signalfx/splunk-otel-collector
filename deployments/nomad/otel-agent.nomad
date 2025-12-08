job "otel-agent" {
  datacenters = ["dc1"]
  type        = "system"

  constraint {
    attribute = "${attr.nomad.version}"
    operator  = "semver"
    value     = "< 1.9.8"
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

      port "jaeger_grpc" {
        to = 14250
      }

      port "jaeger_thrift_http" {
        to = 14268
      }

      port "zipkin" {
        to = 9411
      }

      port "sfx_forwarder" {
        to = 9080
      }

      # Extensions
      port "health_check" {
        to = 13133
      }

      port "zpages" {
        to = 55679
      }
    }

    service {
      name = "otel-agent"
      port = "health_check"
      tags = ["health"]

      check {
        type     = "http"
        port     = "health_check"
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
      port = "jaeger_grpc"
      tags = ["jaeger_grpc"]
    }

    service {
      name = "otel-agent"
      port = "jaeger_thrift_http"
      tags = ["jaeger_thrift_http"]
    }

    service {
      name = "otel-agent"
      port = "zipkin"
      tags = ["zipkin"]
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
      port = "sfx_forwarder"
      tags = ["sfx_forwarder"]
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
          "jaeger_grpc",
          "jaeger_thrift_http",
          "zipkin",
          "health_check",
          "zpages",
          "sfx_forwarder",
        ]
      }

      env {
        SPLUNK_ACCESS_TOKEN = "<SPLUNK_ACCESS_TOKEN>"
        SPLUNK_REALM = "<SPLUNK_REALM>"
        SPLUNK_MEMORY_TOTAL_MIB = 500
      }

      resources {
        cpu    = 100
        memory = 500
      }

      template {
        data        = <<EOF
extensions:
  health_check:
    endpoint: 0.0.0.0:13133
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
  zipkin:
    endpoint: 0.0.0.0:9411
processors:
  batch: null
  memory_limiter:
    check_interval: 2s
    limit_mib: ${SPLUNK_MEMORY_LIMIT_MIB}
  resourcedetection:
    detectors:
    - system
    - env
    override: true
    timeout: 10s
exporters:
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    api_url: https://api.${SPLUNK_REALM}.signalfx.com
    correlation: null
    ingest_url: https://ingest.${SPLUNK_REALM}.signalfx.com
    sync_host_metadata: true
  debug:
    verbosity: detailed
  otlphttp:
    traces_endpoint: "https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace/otlp"
    headers:
      "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
service:
  extensions:
  - health_check
  - zpages
  pipelines:
    metrics:
      exporters:
      - debug
      - signalfx
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - hostmetrics
    metrics/agent:
      exporters:
      - debug
      - signalfx
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - prometheus/agent
    traces:
      exporters:
      - debug
      - otlphttp
      - signalfx
      processors:
      - memory_limiter
      - batch
      - resourcedetection
      receivers:
      - otlp
      - jaeger
      - zipkin
EOF
        destination = "local/config/otel-agent-config.yaml"
      }
    }
  }
}
