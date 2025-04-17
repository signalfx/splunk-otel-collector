# Class for setting the registry values for the splunk-otel-collector service
class splunk_otel_collector::collector_win_registry {
  # Ensure splunk_otel_collector::collector_win_config_options is applied first
  require splunk_otel_collector::collector_win_config_options

  $unordered_collector_env_vars = $splunk_otel_collector::collector_additional_env_vars.map |$var, $value| {
    "${var}=${value}"
  }
  + if !empty($splunk_otel_collector::collector_additional_env_vars) or
  versioncmp($splunk_otel_collector::collector_version, '0.98.0') < 0 {
    $splunk_otel_collector::collector_win_config_options::collector_env_vars.map |$var, $value| {
      "${var}=${value}"
    }
  } else {
    []
  }

  $collector_env_vars = sort($unordered_collector_env_vars)

  if !empty($collector_env_vars) {
    registry_value { "HKLM\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector\\Environment":
      ensure  => 'present',
      type    => array,
      data    => $collector_env_vars,
      require => Registry_key["HKLM\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector"],
    }
  }
}
