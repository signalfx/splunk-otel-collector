# Splunk Forwarder Integration

> WARNING: The following is not supported or tested today. Performance testing,
> sizing guidelines, and latency impact are not provided. Multiline messages
> may not work properly. Other limitations or issues may exist.

Use-case: An existing Splunk Enterprise or Splunk Cloud customer with the
Universal or Heavy Forwarder already deployed wants to send some/all data
from the forwarder to Splunk Log Observer.

The only supported way to collect and send logs to Splunk Log Observer is via
Splunk OpenTelemetry Connector using one of the [recommended
ways](../#getting-started). It is possible, but not supported, to configure a
Splunk Forwarder to send data to Splunk OpenTelemetry Collector in the
following ways:

- Splunk Universal Forwarder with TCP out to Fluentd in Splunk OpenTelemetry
  Collector
- Splunk Heavy Forwarder with either TCP or Syslog out to Fluentd in Splunk
  OpenTelemetry Collector

Please note that there are differences between a Universal and Heavy Forwarder:

- Read
  https://docs.splunk.com/Documentation/SplunkCloud/latest/Forwarding/Typesofforwarders
- Neither the UF nor HF adds Splunk leveraged metadata (e.g. source,
  sourcetype, and host) when forwarding in TCP
- UF supports all-or-nothing forwarding; HF offers flexibility

The most common deployment models to test Splunk Forwarder integration are:

- UF > TCP out > Fluentd (via Splunk OpenTelemetry Connector)
- UF > S2S out > HF > TCP out > Fluentd (via Splunk OpenTelemetry Connector)

## Configuring Universal or Heavy Forwarder

The minimum way to configure with the UF or HF to export data in TCP form is to
add something like the following to `outputs.conf`:

```
[tcpout:fastlane]
server = 127.0.0.1:20001
sendCookedData = false
```

> The `server` configuration above assume the UF/HF and Splunk OpenTelemetry
> Connector are running on the same application. If they are separate, this
> parameter will need to be updated.

Please be advised that advanced configuration may be required to configure
things including TLS, throttling, buffering, parallelism, etc. For more
information, see documentation including:

- [Outputs.conf](https://docs.splunk.com/Documentation/Splunk/8.1.2/Admin/Outputsconf)
- [Multiple pipelinesets](https://docs.splunk.com/Documentation/Forwarder/8.1.2/Forwarder/Configureaforwardertohandlemultiplepipelinesets)

## Configuring Fluentd in Splunk OpenTelemetry Connector

A new source will need to be added to the `conf.d/` directory that Fluentd
monitors. For example, `tcp.conf` could be added with the following:

```
<source>
   @type tcp
   @label @SPLUNK
   tag "tcp.events"
   port 20001
   bind "0.0.0.0"
   delimiter "\n"
   <parse>
     @type "none"
     message_key "_raw"
   </parse>
 </source>
```

> The `@label @SPLUNK` configuration is critical as by default Fluentd only
> collects events with this label.

Multiline events are likely to be an issue with the example configuration
above. Consider configuring the Fluentd [multiline
parser](https://docs.fluentd.org/parser/multiline) to address this.
