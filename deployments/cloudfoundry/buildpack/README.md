# OpenTelemetry Pivotal Cloud Foundry (PCF) Buildpack

A [Cloud Foundry buildpack](https://docs.pivotal.io/application-service/2-11/buildpacks/) to install
the OpenTelemetry Collector for use with PCF apps.  This will probably work for generic Cloud Foundry
apps as well, but it is only tested and supported on Pivotal Platform.

The buildpack's default functionality, as described in this document, is to deploy the OpenTelemetry Collector
as a sidecar for the given app that's being deployed. The Collector is able to observe the app as a nozzle through
the Loggregator Firehose. The Loggregator Firehose is one of the architectures Cloud Foundry
uses to emit logs and metrics. This means that the OpenTelemetry Collector will be observing all
apps and deployments that emit metrics and logs to the Loggregator Firehose as long as it's running.

## Installation
- Clone this repository.
- Change to this directory.
- Run the following command:
```sh
# Add buildpack for OpenTelemetry Collector
$ cf create-buildpack otel_collector_buildpack . 99 --enable
```

### Using PCF Buildpack With an Application
This section covers basic cf CLI (Cloud Foundry Command Line Interface) commands to use the buildpack. 
```sh
# Basic setup, see the configuration section for more envvars that can be set
$ cf set-env <app-name> OTEL_CONFIG <config_file_name>
$ cf set-env <app-name> OTELCOL <desired_collector_executable_name>

$ cd <my-application-direcory>/

# How to run an application without a manifest.yml file:
# Note: This will provide the Collector for the app, but the Collector will not be running.
$ cf push <app-name> -b otel_collector_buildpack -b <main_buildpack>

# How to run an application with a manifest.yml file:
# Option 1)
$ cf push
# Option 2)
$ cf push <app-name> -b otel_collector_buildpack -b <main_buildpack> -f manifest.yml
```

Note: This buildpack requires another buildpack to be supplied after it, it is not allowed to
be the last one for an app. Also, the manifest.yml file will need to provide the
command to run the OpenTelemetry Collector as a sidecar for the application.

## Configuration

The following only applies if you are using the `otelconfig.yaml` config
provided by the buildpack.  If you provide a custom configuration file for the OpenTelemetry Collector 
in your application (and refer to it in the sidecar configuration), these might not
work unless you have preserved the references to the environment variables in the config file.
For proper functionality, the `OTEL_CONFIG` environment variable must point to
the configuration file, whether using the default or a customer version.

Set the following environment variables with `cf set-env` as applicable to configure this buildpack, or
include them in the `manifest.yml` file, as shown in the [included example](#sidecar-configuration).

Required:
- `RLP_GATEWAY_ENDPOINT` - The URL of the RLP gateway that acts as the proxy for the firehose,
    e.g. `https://log-stream.sys.<TAS environment name>.cf-app.com`
- `UAA_ENDPOINT` - The URL of UAA provider,
    e.g. `https://uaa.sys.<TAS environment name>.cf-app.com`
- `UAA_USERNAME` - Name of the UAA user.
- `UAA_PASSWORD` - Password for the UAA user.
- `SIGNALFX_ACCESS_TOKEN` - Your SignalFx organization access token.

Optional:
- `OS` - Operating system that Cloud Foundry is running. Must match format of Otel Collector executable name.
    Default: `linux_amd64`
- `OTEL_CONFIG` - Local name of OpenTelemetry config file. Default: `otelconfig.yaml`
- `OTEL_VERSION` - Executable version of OpenTelemetry Collector (contrib) to use. The buildpack depends on features present in version
    0.47.0+. Default: `0.47.0`
- `OTELCOL` - OpenTelemetry Collector executable file name. Default: `otelcontribcol_$OS-v$OTEL_VERSION`
- `OTELCOLLECTOR_DOWNLOAD_URL` - URL to download the OpenTelemetry Collector from.

- `RLP_GATEWAY_SHARD_ID` - Metrics are load balanced between receivers that use the same shard ID.
   Only use if multiple receivers must receive all metrics instead of
   balancing metrics between them. Default: `otelcol`
- `RLP_GATEWAY_TLS_INSECURE` - Whether to skip TLS verify for the RLP gateway endpoint. Default: `true`
- `UAA_TLS_INSECURE` - Whether to skip TLS verify for the UAA endpoint. Default: `true`
- `SIGNALFX_INGEST_URL` - The ingest base URL for SignalFx. If specified, SIGNALFX_API_URL is also required.
   This option takes precedence over SIGNALFX_REALM.
- `SIGNALFX_API_URL` - The API server base URL for SignalFx. If specified, SIGNALFX_INGEST_URL is also required.
   This option take precedence over SIGNALFX_REALM.
- `SIGNALFX_REALM` - The SignalFx realm in which your organization resides. Used to derive SIGNALFX_INGEST_URL and
   SIGNALFX_API_URL.

## Sidecar Configuration

The recommended method for running the Collector is to run it as a sidecar using
the Cloud Foundry [sidecar
functionality](https://docs.cloudfoundry.org/devguide/sidecars.html).
Additional information can be found [in the v3 API
docs](http://v3-apidocs.cloudfoundry.org/version/release-candidate/#sidecars).

Here is an example application `manifest.yml` file that would run the Collector as
a sidecar:

```yaml
---
applications:
  - name: test-app
    buildpacks:
      - otel_collector_buildpack
      - go_buildpack
    instances: 1
    memory: 256M
    random-route: true
    env:
      RLP_GATEWAY_ENDPOINT: "https://log-stream.sys.<TAS environment name>.cf-app.com"
      UAA_ENDPOINT: "https://uaa.sys.<TAS environment name>.cf-app.com"
      UAA_USERNAME: "..."
      UAA_PASSWORD: "..."
      SIGNALFX_ACCESS_TOKEN: "..."
      SIGNALFX_REALM: "..."
    sidecars:
      - name: otel-collector
        process_types:
          - web
        command: "$HOME/../deps/otelcontribcol_${OS:-linux_amd64}-v${OTEL_VERSION:-0.47.0} --config=$HOME/../deps/${OTEL_CONFIG:-otelconfig.yaml}"
        memory: 100MB
```
If using a `manifest.yaml` file, you may push your app simply with the following command:
```sh
# If you are using cf CLI v7
$ cf push

# If you are using cf CLI v6
$ cf v3-push <app-name>
```
This will deploy the app with the proper buildpacks, and the OpenTelemetry Collector running in the sidecar configuration.

## Troubleshooting

* If the app is running but the OpenTelemetry Collector is not, it may be that the sidecar configuration is not
being picked up properly from the manifest file. Try running the following commands:

```sh
# This will apply the manifest file to an existing app
$ cf v3-apply-manifest -f manifest.yml
# This will re-load the app with the sidecar configuration included
$ cf push
```

* Another possibility is the sidecar was not allocated memory. The `memory` option
is required for a sidecar process for it to run. Once memory allocation has been added to the sidecar,
re-run the above command to apply the manifest and push the application again.

### Useful CF CLI debugging commands
```sh
$ cf apps # Checks status of app
$ cf logs <app-name> --recent # View the app's logs
$ cf env <app-name> # Show all environment variables for the app.
$ cf events <app-name> # View the app's events
```