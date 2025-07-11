# Class for collecting the configuration options for the splunk-otel-collector service
class splunk_otel_collector::collector_win_config_options {
    $base_env_vars = {
        'SPLUNK_ACCESS_TOKEN' => $splunk_otel_collector::splunk_access_token,
        'SPLUNK_API_URL' => $splunk_otel_collector::splunk_api_url,
        'SPLUNK_BUNDLE_DIR' => $splunk_otel_collector::splunk_bundle_dir,
        'SPLUNK_COLLECTD_DIR' => $splunk_otel_collector::splunk_collectd_dir,
        'SPLUNK_CONFIG' => $splunk_otel_collector::collector_config_dest,
        'SPLUNK_HEC_TOKEN' => $splunk_otel_collector::splunk_hec_token,
        'SPLUNK_HEC_URL' => $splunk_otel_collector::splunk_hec_url,
        'SPLUNK_INGEST_URL' => $splunk_otel_collector::splunk_ingest_url,
        'SPLUNK_MEMORY_TOTAL_MIB' => $splunk_otel_collector::splunk_memory_total_mib,
        'SPLUNK_REALM' => $splunk_otel_collector::splunk_realm,
    }

    $gomemlimit = if ($splunk_otel_collector::collector_version == 'latest' or
        versioncmp($splunk_otel_collector::collector_version, '0.97.0') >= 0) and
        !$splunk_otel_collector::gomemlimit.strip().empty() {
            { 'GOMEMLIMIT' => $splunk_otel_collector::gomemlimit }
        } else {
            {}
        }

    $listen_interface = if !$splunk_otel_collector::splunk_listen_interface.strip().empty() {
            { 'SPLUNK_LISTEN_INTERFACE' => $splunk_otel_collector::splunk_listen_interface }
        } else {
            {}
        }

    $custom_cmd_line = if !$splunk_otel_collector::collector_command_line_args.strip().empty() and
        versioncmp($splunk_otel_collector::collector_version, '0.127.0') >= 0 {
            { 'COLLECTOR_SVC_ARGS' => $splunk_otel_collector::collector_command_line_args }
        } else {
            {}
        }

    $collector_env_vars = $base_env_vars + $gomemlimit + $listen_interface + $custom_cmd_line
}
