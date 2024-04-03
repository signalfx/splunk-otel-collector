# Changelog

## ansible-v0.25.0

### 💡 Enhancements 💡

- On Windows the `SPLUNK_*` environment variables were moved from the machine scope to the collector service scope.
  It is possible that some instrumentations are relying on the machine-wide environment variables set by the installation. ([#3930](https://github.com/signalfx/splunk-otel-collector/pull/3930))

### 🧰 Bug fixes 🧰

- Use more secure assert calls ([#4024](https://github.com/signalfx/splunk-otel-collector/pull/4024))


## ansible-v0.25.0

### 💡 Enhancements 💡

- Initial support for [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js) Auto Instrumentation on Linux

## ansible-v0.24.0

### 🚩 Deprecations 🚩

- The `splunk_otel_auto_instrumentation_generate_service_name` and `splunk_otel_auto_instrumentation_disable_telemetry`
  options are deprecated and only applicable if `splunk_otel_auto_instrumentation_version` is < `0.87.0`.

### 💡 Enhancements 💡

- Support Splunk OpenTelemetry Auto Instrumentation [v0.87.0](
  https://github.com/signalfx/splunk-otel-collector/releases/tag/v0.87.0) and newer (Java only).
- Support activation and configuration of auto instrumentation for only `systemd` services.
- Support setting the OTLP exporter endpoint for auto instrumentation (default: `http://127.0.0.1:4317`). Only
  applicable if `splunk_otel_auto_instrumentation_version` is `latest` or >= `0.87.0`.

### 🧰 Bug fixes 🧰

- Ensure loop variables for Windows tasks are defined

## ansible-v0.23.0

### 💡 Enhancements 💡

- Only propagate `splunk_listen_interface` to target SPLUNK_LISTEN_INTERFACE service environment variable if set.

## ansible-v0.22.0

### 💡 Enhancements 💡

- Support the `splunk_hec_token` option for Linux (thanks @chepati)

## ansible-v0.21.0

### 💡 Enhancements 💡

- Add support for the `splunk_listen_interface` and `signalfx_dotnet_auto_instrumentation_global_tags` options

### 🧰 Bug fixes 🧰

- Update default base url for downloading `td-agent` on Windows to `https://s3.amazonaws.com/packages.treasuredata.com`

## ansible-v0.20.0

### 🛑 Breaking changes 🛑

- Fluentd installation is ***disabled*** by default.
  - Specify the `install_fluentd: yes` role variable in your playbook to enable installation.

## ansible-v0.19.0

### 💡 Enhancements 💡

- Add support for Auto Instrumentation for .NET on Windows

## ansible-v0.18.0

### 💡 Enhancements 💡

- Add support for RHEL 9 and Amazon Linux 2023

## ansible-v0.17.0

### 🧰 Bug fixes 🧰

- Fix Windows templating [(#2960)](https://github.com/signalfx/splunk-otel-collector/pull/2960)

## ansible-v0.16.0

### 💡 Enhancements 💡

- Add support for additional options for Splunk OpenTelemetry Auto Instrumentation for Java (Linux only)

## ansible-v0.15.0

### 💡 Enhancements 💡

- Add `splunk_otel_collector_additional_env_vars` option to allow passing additional environment variables to the collector service

### 🧰 Bug fixes 🧰

- Fix custom URLs and HEC token configuration on windows

## ansible-v0.14.0

### 💡 Enhancements 💡

- Add `splunk_otel_collector_no_proxy` var to update service NO_PROXY environment variable (Linux only) [(#2482)](https://github.com/signalfx/splunk-otel-collector/pull/2482)

## ansible-v0.13.0

### 💡 Enhancements 💡

- Add `splunk_skip_repo` var to disable adding Splunk deb/rpm repos (#2249)

## ansible-v0.12.0

### 💡 Enhancements 💡

- Support installing with ansible and skipping restart of services (#1930)

## ansible-v0.11.0

### 💡 Enhancements 💡

- Support downloading the Splunk Collector Agent and fluentd using a proxy on Windows.

## ansible-v0.10.0

### 💡 Enhancements 💡

- Update default `td-agent` version to 4.3.2 to support log collection with fluentd on Ubuntu 22.04

## ansible-v0.9.0

### 💡 Enhancements 💡

- Initial support for Splunk OpenTelemetry Auto Instrumentation for Java (Linux only)

## ansible-v0.8.0

### 💡 Enhancements 💡

- Add support for Ubuntu 22.04 (collector only; log collection with Fluentd [not currently supported](https://www.fluentd.org/blog/td-agent-v4.3.1-has-been-released))

## ansible-v0.7.0

### 🛑 Breaking changes 🛑

- Removed support for Debian 8
- Bump minimum supported ansible version to 2.10

### 💡 Enhancements 💡

- Add support for Debian 11

## ansible-v0.6.0

### 🧰 Bug fixes 🧰

- Fix `SPLUNK_TRACE_URL` and `SPLUNK_HEC_URL` env vars (#1290)

## ansible-v0.5.0

### 💡 Enhancements 💡

- Bump default td-agent version to 4.3.0 (#1218)

### 🧰 Bug fixes 🧰

- Add `ansible.windows` as a galaxy dependency (#1229)

## ansible-v0.4.0

### 💡 Enhancements 💡

- Add an option to provide a custom configuration that will be merged into the
  default one using `splunk_config_override` var (#950)

## ansible-v0.3.0

### 🛑 Breaking changes 🛑

- Rename `agent_bundle_dir` parameter to `splunk_bundle_dir` (#810)

### 💡 Enhancements 💡

- Add Windows support (#646, #797)
- Add SUSE support (collector only) (#748)
- Use `template` instead of `copy` to manage the configuration file (#770)

## ansible-v0.2.0

- Add proxy support for otel collector service runtime level (#699)
- Drop `apt_key` deprecated module in favor of existing `get_url` task (#698)

## ansible-v0.1.3

- Add meta/runtime.yml (#690)

## ansible-v0.1.2

- Install libcap on RHEL platforms (#678)

## ansible-v0.1.1

### 🧰 Bug fixes 🧰

- Remove redundant import_tasks (#420)
- Fill fully qualified task names where missed (#420)
- Fix conditions to avoid warning (#420)

## ansible-v0.1.0

Initial version of `signalfx.splunk_otel_collector` Ansible Collection with
Linux support.
