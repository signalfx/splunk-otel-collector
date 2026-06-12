# Class for collecting the configuration options for the splunk-otel-collector service
class splunk_otel_collector::collector_win_config_options {
    $collector_command_line_configurable = (
        $splunk_otel_collector::collector_version == 'latest' or
        versioncmp($splunk_otel_collector::collector_version, '0.127.0') >= 0
    )
    $custom_config_dest = $splunk_otel_collector::collector_config_dest != $splunk_otel_collector::params::collector_config_dest

    $base_env_vars = {
        'SPLUNK_ACCESS_TOKEN' => $splunk_otel_collector::splunk_access_token,
        'SPLUNK_API_URL' => $splunk_otel_collector::splunk_api_url,
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

    $config_env_var = if $collector_command_line_configurable and !$custom_config_dest {
            {}
        } else {
            { 'SPLUNK_CONFIG' => $splunk_otel_collector::collector_config_dest }
        }

    $collector_config_arg = '--config "' + $splunk_otel_collector::collector_config_dest + '"'
    $collector_command_line_args = if $collector_command_line_configurable and $custom_config_dest {
            if !$splunk_otel_collector::collector_command_line_args.strip().empty() {
                "${splunk_otel_collector::collector_command_line_args} ${collector_config_arg}"
            } else {
                $collector_config_arg
            }
        } elsif $collector_command_line_configurable {
            $splunk_otel_collector::collector_command_line_args
        } else {
            ''
        }

    $collector_command_line_install_options = if $collector_command_line_configurable and
        !$collector_command_line_args.strip().empty() {
            { 'COLLECTOR_SVC_ARGS' => $collector_command_line_args }
        } else {
            {}
        }

    $collector_env_vars = $base_env_vars + $config_env_var + $gomemlimit + $listen_interface
    $collector_install_options = $collector_env_vars + $collector_command_line_install_options
}
