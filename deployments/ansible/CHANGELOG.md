# Changelog

## Unreleased

- Add SUSE support (collector only) (#748)

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
