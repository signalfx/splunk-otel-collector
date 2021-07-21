# Class for setting the registry values for the splunk-otel-collector service
class splunk_otel_collector::collector_win_registry () {
  $registry_key = 'HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment'

  registry_key { $registry_key:
    ensure => 'present',
  }

  registry_value { "${registry_key}\\SPLUNK_ACCESS_TOKEN":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_access_token,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_API_URL":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_api_url,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_BALLAST_SIZE_MIB":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_ballast_size_mib,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_BUNDLE_DIR":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_bundle_dir,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_COLLECTD_DIR":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_collectd_dir,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_CONFIG":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::collector_config_dest,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_HEC_TOKEN":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_hec_token,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_HEC_URL":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_hec_url,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_INGEST_URL":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_ingest_url,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_MEMORY_TOTAL_MIB":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_memory_total_mib,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_REALM":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_realm,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\SPLUNK_TRACE_URL":
    ensure  => 'present',
    type    => 'string',
    data    => $splunk_otel_collector::splunk_trace_url,
    require => Registry_key[$registry_key],
  }
}
