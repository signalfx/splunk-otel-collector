# Splunk OpenTelemetry Collector Ansible Collection

## Description

The `signalfx.splunk_otel_collector` Ansible Collection can be found in the
[signalfx/splunk_otel_collector](./signalfx/splunk_otel_collector) directory, which matches the required
directory hierarchy of `<namespace>/<collection_name>`.

## Certified Collection

To build the certified collection, run the following command from this directory:
```
$ ./build_cisco_collection.sh
```
The certified collection will be located under the local path `./dist/` as a `tar.gz` file.