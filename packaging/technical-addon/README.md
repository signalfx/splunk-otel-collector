# Configuration
## environment variables
In addition to the typical collector environment variables, we additionally provide

- `$SPLUNK_HOME` ex `/opt/splunk`
- `$SPLUNK_OTEL_TA_HOME` ex `/opt/splunk/etc/apps/Splunk_TA_Otel`
- `$SPLUNK_OTEL_TA_PLATFORM_HOME` ex `/opt/splunk/etc/apps/Splunk_TA_Otel/linux_x86_64`

These are useful for setting things in `inputs.conf` such as 
- `splunk_bundle_dir=$SPLUNK_OTEL_TA_PLATFORM_HOME/bin/agent-bundle`
- `splunk_collectd_dir=$SPLUNK_OTEL_TA_PLATFORM_HOME/bin/agent-bundle/run/collectd`

There is a debug environment variable for the TA scripts, which will log verbose
messages if set to anything other than the empty string.
- `SPLUNK_OTEL_TA_DEBUG`

## parameters
See `Splunk_TA_Otel/README/inputs.conf.spec` or `Splunk_TA_Otel/default/inputs.conf` for configuration values.

Of note is that we need to read the `access_token` from a file into an environment variable during TA initialization.
By default, this access_token is expected to live in `$SPLUNK_OTEL_TA_HOME/local/access_token`


# Installation

## Reducing the size of the TA
Customers may want to remove unnecessary configuration if they're only using one
operating system (ex windows, linux).

This can be accomplished by removing either `Splunk_TA_Otel/linux_x86_64` or
`Splunk_TA_Otel/windows_x86_64`, respectively.

Further, they may remove the agent bundle downloaded to the `bin/` folder in these platform specific directories if they don't need smart agent support.

## Maintaining configuration between upgrades
As with all TAs, any changes made to `configs` will be overwritten.
Customers should copy any relevant custom configuration from `configs/` or `defaults/`
to `local/`.

# Autoinstrumentation
For all, you may run `build-linux-autoinstrumentation-ta`.
For specific targets, you may run
1. `make generate-technical-addon-linux-autoinstrumentation` (makes autoinstrumentation and dependencies)
2. `make gen-modinput-config && make build-ta-runners`
