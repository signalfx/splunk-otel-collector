[Splunk_TA_otel://<name>]

splunk_access_token_file = <value>
* File whose contents store the credentials to be set in `SPLUNK_ACCESS_TOKEN` (used to auth with Splunk Observability Cloud, default `$SPLUNK_OTEL_TA_HOME/local/access_token`).

# Below are all "pass through" configuration options that will enable environment variables supported in 
# https://github.com/signalfx/splunk-otel-collector/blob/main/internal/settings/settings.go#L37-L64
gomemlimit = <value>
* Value to use for the `GOMEMLIMIT` environment variable

splunk_api_url = <value>
* Value to use for the `SPLUNK_API_URL` environment variable

splunk_bundle_dir = <value>
* Value to use for the `SPLUNK_BUNDLE_DIR` environment variable (used in smart agent config, default `$SPLUNK_OTEL_TA_PLATFORM_HOME/bin/agent-bundle`).  NOTE: Once extracted, please do not move the folder structure, as some "first-run" patching occurs that depends on the directory it was run in.  Instead, allow the TA to re-extract to the new location if changed.

splunk_collectd_dir = <value>
* Value to use for the `SPLUNK_COLLECTD_DIR` environment variable (used in smart agent config, default `$SPLUNK_OTEL_TA_HOME/bin/agent-bundle/run/collectd`)

splunk_config = <value>
* Value to use for the `SPLUNK_CONFIG` environment variable (default `$SPLUNK_OTEL_TA_HOME/config/ta_agent_config.yaml`)

splunk_config_dir = <value>
* Value to use for the `SPLUNK_CONFIG_DIR` environment variable (default `$SPLUNK_OTEL_TA_HOME/config/`).  Same as `--configd` parameter as defined in https://github.com/signalfx/splunk-otel-collector/blob/v0.114.0/internal/confmapprovider/discovery/README.md?plain=1#L47

splunk_debug_config_server = <value>
* Value to use for the `SPLUNK_DEBUG_CONFIG_SERVER` environment variable

splunk_config_yaml = <value>
* Value to use for the `SPLUNK_CONFIG_YAML` environment variable

splunk_listen_interface = <value>
* Value to use for the `SPLUNK_LISTEN_INTERFACE` environment variable

splunk_gateway_url = <value>
* Value to use for the `SPLUNK_GATEWAY_URL` environment variable

splunk_memory_limit_mib = <value>
* Value to use for the `SPLUNK_MEMORY_LIMIT_MIB` environment variable

splunk_memory_total_mib = <value>
* Value to use for the `SPLUNK_MEMORY_TOTAL_MIB` environment variable

splunk_ingest_url = <value>
* Endpoint for `SPLUNK_API_URL`

splunk_realm = <value>
* Splunk Observability realm to use for the `SPLUNK_REALM` environment variable (ex us0)

discovery = <value>
* Boolean, if `true` will enable `--discovery`

discovery_properties = <value>
* String, same as `--discovery-properties` parameter.  See [properties reference](https://github.com/signalfx/splunk-otel-collector/blob/v0.114.0/internal/confmapprovider/discovery/README.md?plain=1#L175)

configd = <value>
* Boolean, if `true` will enable `--configd`
