# Class for setting the registry values for the splunk-otel-collector service
class splunk_otel_collector::collector_win_registry () {
  $unordered_collector_env_vars = $splunk_otel_collector::collector_additional_env_vars.map |$var, $value| {
    "${var}=${value}"
  }
  + [
    "SPLUNK_ACCESS_TOKEN=${splunk_otel_collector::splunk_access_token}",
    "SPLUNK_API_URL=${splunk_otel_collector::splunk_api_url}",
    "SPLUNK_BUNDLE_DIR=${splunk_otel_collector::splunk_bundle_dir}",
    "SPLUNK_COLLECTD_DIR=${splunk_otel_collector::splunk_collectd_dir}",
    "SPLUNK_CONFIG=${splunk_otel_collector::collector_config_dest}",
    "SPLUNK_HEC_TOKEN=${splunk_otel_collector::splunk_hec_token}",
    "SPLUNK_HEC_URL=${splunk_otel_collector::splunk_hec_url}",
    "SPLUNK_INGEST_URL=${splunk_otel_collector::splunk_ingest_url}",
    "SPLUNK_MEMORY_TOTAL_MIB=${splunk_otel_collector::splunk_memory_total_mib}",
    "SPLUNK_REALM=${splunk_otel_collector::splunk_realm}",
    "SPLUNK_TRACE_URL=${splunk_otel_collector::splunk_trace_url}",
  ]
  + if !$splunk_otel_collector::splunk_ballast_size_mib.strip().empty() {
    ["SPLUNK_BALLAST_SIZE_MIB=${splunk_otel_collector::splunk_ballast_size_mib}"]
  } else {
    []
  }
  + if !$splunk_otel_collector::splunk_listen_interface.strip().empty() {
    ["SPLUNK_LISTEN_INTERFACE=${splunk_otel_collector::splunk_listen_interface}"]
  } else {
    []
  }

  $collector_env_vars = sort($unordered_collector_env_vars)

  registry_value { "HKLM\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector\\Environment":
    ensure  => 'present',
    type    => array,
    data    => $collector_env_vars,
    require => Registry_key["HKLM\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector"],
  }
}
