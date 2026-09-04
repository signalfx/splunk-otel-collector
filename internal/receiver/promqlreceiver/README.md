# PromQL Receiver

The PromQL receiver periodically executes PromQL queries against a Prometheus HTTP API and converts the query results into OpenTelemetry metrics.

This makes it possible to collect metrics that have already been selected, filtered, or calculated by Prometheus rather than scraping all of the underlying source metrics into the Collector.

> **Status: Development**
>
> This receiver is currently experimental. It is implemented under `internal/receiver/promqlreceiver`, but it is not currently registered in the default public Splunk OpenTelemetry Collector component set.
>
> The implementation also has several known limitations described in [Current limitations](#current-limitations). It should not currently be treated as a production-ready receiver.

The team is looking to donate this exporter to OpenTelemetry. Please see this [issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/50723) to learn more, upvote, and offer to be a codeowner.
---

## Why this receiver exists

The standard Prometheus receiver and the PromQL receiver solve different problems.

### Prometheus receiver

The standard Prometheus receiver retrieves exposed metrics from targets:

```text
Application
│
│ GET /metrics
▼
Prometheus receiver
│
▼
OTel metrics
```

For example, the application might expose:

```text
http_requests_total{service="checkout",status="200"} 15423
http_requests_total{service="checkout",status="500"} 217
```

The Collector receives those source time series.

### PromQL receiver

The PromQL receiver instead talks to a Prometheus server:

```text
Collector
│
│ sum(rate(http_requests_total[5m]))
▼
Prometheus /api/v1/query
│
│ evaluated result
▼
PromQL receiver
│
▼
OTel metrics
```

Prometheus performs the query first. The receiver then turns the returned samples into OpenTelemetry metrics.

This can be useful when the desired telemetry is represented by a PromQL query rather than by a directly scrapeable metric.

Potential use cases include:

- retrieving a selected set of metrics from a central Prometheus server;
- reducing the number of raw time series that must be sent downstream;
- querying metrics that are not directly accessible to the Collector;
- using Prometheus as the query and aggregation layer before exporting results elsewhere;
- eventually representing calculated PromQL expressions as OpenTelemetry metrics.

Some complex PromQL expressions currently interact poorly with metric-type metadata. See [PromQL and metric type metadata](#promql-and-metric-type-metadata).

# Configuration

The receiver configuration combines three sets of settings:

1. PromQL queries;
2. standard Collector HTTP client configuration;
3. standard Collector scraper-controller configuration.

The component-specific configuration is defined in [config.go](./config.go).

A minimal configuration looks like this:

```yaml
receivers:
  promql:
endpoint: http://localhost:9090/api/v1/query
queries:
- up

exporters:
  debug:
verbosity: detailed

service:
  pipelines:
metrics:
  receivers:
  - promql
  exporters:
  - debug
```

---

# HTTP client configuration

The receiver uses the Collector's standard HTTP client configuration.

This means HTTP client options such as TLS, headers, proxy configuration, authentication extensions, connection settings, and client timeouts can be used where supported by the Collector version containing the receiver.

For example:

```yaml
receivers:
  promql:
endpoint: https://prometheus.example.com/api/v1/query
timeout: 10s
headers:
  X-Scope-OrgID: example-tenant
queries:
- up
```

## TLS

Example:

```yaml
receivers:
  promql:
endpoint: https://prometheus.example.com/api/v1/query
tls:
  ca_file: /etc/otel/certs/ca.pem
queries:
- up
```

# Query execution

For each configured query, the receiver performs an HTTP `GET`.

Given:

```yaml
endpoint: http://prometheus:9090/api/v1/query

queries:
- up
```

the resulting request is equivalent to:

```text
GET /api/v1/query?query=up
```

PromQL is URL encoded automatically.

For example:

```yaml
queries:
- 'up{job="checkout"}'
```

becomes conceptually:

```text
GET /api/v1/query?query=up%7Bjob%3D%22checkout%22%7D
```

The receiver does not currently specify an explicit Prometheus query `time` parameter, so the Prometheus server evaluates the instant query using its normal default evaluation time.

---

# Prometheus response handling

The receiver expects the standard Prometheus API response structure:

```json
{
"status": "success",
"data": {
"resultType": "vector",
"result": []
}
}
```

The receiver also parses the following top-level response fields:

```text
status
data
errorType
error
warnings
infos
```

Currently, `warnings` and `infos` are parsed but are not propagated or logged.

If the Prometheus response reports:

```json
{
"status": "error"
}
```

the scrape returns an error containing the Prometheus error message.

---

## Metric names

When the Prometheus result contains a `__name__` label, that value becomes the OpenTelemetry metric name.

For example:

```text
__name__="up"
```

becomes:

```text
Metric name: up
```

### Queries that remove the metric name

PromQL operations can remove the original `__name__` label.

When that happens, the receiver uses:

```text
promql_result
```

as the metric name.

For example, a query such as:

```promql
sum(up)
```

may produce a result without `__name__`.

The receiver then creates:

```text
promql_result
```

as the OpenTelemetry metric name.

This is currently a fixed fallback name and is not configurable.

---

# Label conversion

Prometheus labels are converted into OpenTelemetry data-point attributes.

For example:

```text
up{
__name__="up",
__type__="gauge",
job="prometheus",
instance="localhost:9090"
}
```

becomes conceptually:

```text
Metric
name: up
type: Gauge

DataPoint
attributes:
job: prometheus
instance: localhost:9090
```

The following labels are intentionally excluded from attributes:

```text
__name__
__type__
```

because they are used to determine the OpenTelemetry metric name and type.

Other labels are copied directly to the data point.

---

# Prometheus metric types

The receiver uses the Prometheus `__type__` label to determine which OpenTelemetry metric type to create.

Current conversion behavior is:

| Prometheus `__type__` | OpenTelemetry type |
|---|---|
| `gauge` | Gauge |
| `counter` | Sum |
| `histogram` | Histogram |
| `gaugehistogram` | Histogram |

Other types are not currently handled.

---

## Gauge conversion

A Prometheus gauge becomes an OpenTelemetry Gauge.

Conceptually:

```text
temperature_celsius{
__type__="gauge",
room="server-room"
} 24.5
```

becomes:

```text
Metric:
name: temperature_celsius
type: Gauge

DataPoint:
value: 24.5
attributes:
room: server-room
```

---

## Counter conversion

A Prometheus counter becomes an OpenTelemetry Sum.

Conceptually:

```text
requests_total{
__type__="counter",
service="checkout"
} 14525
```

becomes:

```text
Metric:
name: requests_total
type: Sum

DataPoint:
value: 14525
attributes:
service: checkout
```

The current implementation does **not** explicitly set:

- aggregation temporality; or
- monotonicity

on the resulting OpenTelemetry Sum.

That needs to be addressed for fully correct counter semantics.

---

## Histogram conversion

Prometheus histogram and gauge-histogram samples are converted into OpenTelemetry Histogram metrics.

The receiver currently copies:

- timestamp;
- sum;
- count;
- bucket counts;
- labels as attributes.

The current implementation does not populate explicit OpenTelemetry bucket bounds while appending the histogram bucket counts.

Histogram conversion should therefore be considered incomplete and needs additional validation before production use.

---

# PromQL and metric type metadata

The receiver currently depends heavily on the Prometheus:

```text
__type__
```

label.

Without it, `convertVector()` cannot determine whether a float sample should become an OpenTelemetry Gauge or Sum.

Recent Prometheus versions can expose type and unit metadata in PromQL through the experimental:

```text
type-and-unit-labels
```

feature.

This causes Prometheus to inject reserved labels including:

```text
__type__
__unit__
```

from metric metadata.

A Prometheus server may therefore need to run with the equivalent of:

```text
--enable-feature=type-and-unit-labels
```

for the current implementation to receive the type metadata it expects.

## PromQL can remove type metadata

There is an important complication.

Prometheus treats `__type__` and `__unit__` similarly to `__name__` in many PromQL operations.

PromQL expressions may therefore remove these labels.

For example, calculations and aggregation expressions may produce a perfectly valid numeric result while no longer returning:

```text
__type__
```

The receiver contains a fallback for a missing `__name__`:

```text
promql_result
```

but currently has **no equivalent fallback or configuration for a missing `__type__`**.

This means support for arbitrary calculated PromQL expressions is not yet complete.

A future implementation should define how the OpenTelemetry output metric type is determined for derived PromQL expressions.

Possible approaches include:

- configuring the expected metric type with each query;
- deriving the type from Prometheus metadata when available;
- assigning a documented default;
- returning a configuration or scrape error when type information is unavailable.

---

# Unit metadata

Prometheus can also provide:

```text
__unit__
```

metadata.

The current receiver does not map `__unit__` to the OpenTelemetry metric `unit` field.

Unlike `__name__` and `__type__`, the `__unit__` label is not filtered by the current conversion code, so if present it is currently treated as an ordinary data-point attribute.

This is another area where the Prometheus-to-OpenTelemetry conversion can be improved.

---

# Resource and scope organization

Each successful PromQL query currently creates its own `ResourceMetrics` entry.

Conceptually:

```text
Scrape
├── Query 1 result
│   └── ResourceMetrics
│       └── ScopeMetrics
│           └── Metrics
│
├── Query 2 result
│   └── ResourceMetrics
│       └── ScopeMetrics
│           └── Metrics
│
└── Query 3 result
└── ResourceMetrics
└── ScopeMetrics
└── Metrics
```

No resource attributes are currently added by the PromQL receiver itself.

The receiver also does not currently set an instrumentation scope name or version on the produced `ScopeMetrics`.

Resource and metric processors can still be used later in the Collector pipeline.

---

# Query failure behavior

`ScrapeMetrics()` attempts every configured query and collects errors using `errors.Join()`.

Conceptually:

```go
for _, query := range queries {
if err := runQuery(query); err != nil {
errors = append(errors, err)
}
}
```

This means one failing query does not prevent the receiver from attempting subsequent queries.

However, the returned error is currently a normal scrape error rather than an OpenTelemetry partial-scrape error.

With the standard scraper helper, a normal scrape error can cause the metrics returned by that scrape to be discarded rather than forwarded to the next consumer.

As a result, the practical behavior should currently be treated as:

> **One failed query can cause the entire collection cycle to fail.**

If partial query success is desired, query failures should be represented using the Collector's partial-scrape error mechanism.

---

# HTTP error behavior

The receiver:

1. performs the HTTP request;
2. reads the response body;
3. unmarshals the Prometheus JSON response;
4. checks the Prometheus API `status` field.

It does not currently perform an explicit HTTP status-code check before decoding the body.

Therefore, a non-Prometheus error page or proxy response may appear as a JSON decoding error rather than a more specific HTTP error.

---
# Example: multiple queries
```yaml
receivers:
  promql:
    endpoint: http://prometheus:9090/api/v1/query
    collection_interval: 30s
    queries:
    - up
    - process_cpu_seconds_total
    - go_goroutines
processors:
  batch:
exporters:
  debug:
    verbosity: detailed
service:
  pipelines:
    metrics:
      receivers:
      - promql
      processors:
      - batch
      exporters:
      - debug
```
Every 30 seconds the receiver executes the three queries sequentially and attempts to convert each returned instant vector into OpenTelemetry metrics.

---

# Intended derived-query example
A major reason for a PromQL receiver is to retrieve calculated metrics.
For example:
```promql
sum by (service) (
  rate(http_requests_total{status=~"5.."}[5m])
)
```

Conceptually, Prometheus could calculate the error rate and return:
```text
  {service="checkout"} 1.72
  {service="payments"} 0.41
```
instead of requiring the Collector to ingest all of the underlying `http_requests_total` time series.

The intended flow is:

```text
raw time series
│
▼
Prometheus
│
│ evaluate:
│ sum by (service) (
│   rate(http_requests_total{status=~"5.."}[5m])
│ )
▼
calculated vector
│
▼
PromQL receiver
│
▼
OTel metric
```
