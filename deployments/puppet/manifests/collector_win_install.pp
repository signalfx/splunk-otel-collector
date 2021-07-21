# Download and install the splunk-otel-collector MSI on Windows
class splunk_otel_collector::collector_win_install ($repo_url, $version, $package_name, $service_name) {
  $msi_name = "splunk-otel-collector-${version}-amd64.msi"
  $collector_path = "${::win_programfiles}\\Splunk\\OpenTelemetry Collector\\otelcol.exe"
  $registry_key = 'HKLM\SYSTEM\CurrentControlSet\Services\splunk-otel-collector'

  # Only download and install if not already installed or version does not match
  if $::win_collector_path != $collector_path or $::win_collector_version != $version {
    file { "${::win_temp}\\${msi_name}":
      source => "${repo_url}/${msi_name}"
    }

    -> package { $package_name:
      ensure => $version,
      source => "${::win_temp}\\${msi_name}",
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
