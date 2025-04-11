# Download and install the splunk-otel-collector MSI on Windows
class splunk_otel_collector::collector_win_install ($repo_url, $version, $package_name, $service_name) {
  contain 'splunk_otel_collector::collector_win_config_options'

  $msi_name = "splunk-otel-collector-${version}-amd64.msi"
  $collector_path = "${facts['win_programfiles']}\\Splunk\\OpenTelemetry Collector\\otelcol.exe"
  $registry_key = 'HKLM\SYSTEM\CurrentControlSet\Services\splunk-otel-collector'

  # Only download and install if not already installed or version does not match
  if $facts['win_collector_path'] != $collector_path or $facts['win_collector_version'] != $version {
    file { "${facts['win_temp']}\\${msi_name}":
      source => "${repo_url}/${msi_name}",
    }

    -> package { $package_name:
      source          => "${facts['win_temp']}\\${msi_name}",
      require         => Class['splunk_otel_collector::collector_win_config_options'],
      # If the MSI is not configurable, the install_options below will be ignored during installation.
      install_options => $splunk_otel_collector::collector_win_config_options::collector_env_vars,
    }
  }

  # Ensure the registry values are always up-to-date
  registry_key { $registry_key:
    ensure => 'present',
  }

  registry_value { "${registry_key}\\ExePath":
    ensure  => 'present',
    type    => 'string',
    data    => $collector_path,
    require => Registry_key[$registry_key],
  }

  registry_value { "${registry_key}\\CurrentVersion":
    ensure  => 'present',
    type    => 'string',
    data    => $version,
    require => Registry_key[$registry_key],
  }
}
