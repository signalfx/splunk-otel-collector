# Partial reload

## Scope

The Universal Forwarder (UF) reload model is input-oriented: a change to one
TA reloads the affected input stanzas, while a system-layer change may affect
all TAs. Rebuilding every input in a TA is still broader than necessary because
an `inputs.conf` file can contain several independent stanzas.

There are three practical approaches:

1. Rebuild the complete Collector graph. This is the existing safe fallback,
   but it loses component state and creates unnecessary downtime.
2. Use the OpenTelemetry Collector service partial-reload feature. With
   `service.partialReload` and `service.partialReloadReceivers` enabled, the
   Collector diffs its effective configuration and updates only changed
   receiver components. Non-receiver changes still use a full reload.
3. Reconcile components owned by the TA observer. This is required for
   `splunk_inputs`, because TA files are outside the Collector configuration
   graph. The observer now diffs each effective input stanza, including the
   `props.conf` and `transforms.conf` used to build its operator pipeline.

The implementation uses both approaches: the upstream service handles normal
Collector configuration changes, and `splunk_inputs` handles TA-local changes.
An unchanged stanza retains its running receiver; only added, removed, or
changed stanzas are stopped and started.

## Explicit local reload

Code that already knows a local TA configuration changed can trigger an
immediate reconciliation with the package function:

```go
err := splunkinputsreceiver.Reload(ctx, rcvr)
```

The built-in filesystem watcher continues to debounce and invoke targeted
reconciliation automatically.

## Boundaries

Changing a `props.conf` or `transforms.conf` layer rebuilds every affected TA
input because those files are compiled into each input's operator pipeline.
Changing only an output or another service-level setting remains a full
Collector reload unless the owning component provides its own reload contract.

This follows the OpenTelemetry partial-reload RFC's rule that changes are
propagated only to the affected component boundary and that unsupported change
types fall back to a full reload:
<https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/partial-reload.md>
