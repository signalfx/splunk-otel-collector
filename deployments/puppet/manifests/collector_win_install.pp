# Download and install the splunk-otel-collector MSI on Windows
#
# @param repo_url
# @param version
# @param package_name
# @param service_name
#
class splunk_otel_collector::collector_win_install (
  String $repo_url,
  String $version,
  String $package_name,
  String $service_name,
) {
  $msi_name = "splunk-otel-collector-${version}-amd64.msi"
  $collector_path = "${facts['win_programfiles']}\\Splunk\\OpenTelemetry Collector\\otelcol.exe"
  $registry_key = 'HKLM\SYSTEM\CurrentControlSet\Services\splunk-otel-collector'

  # Only download and install if not already installed or version does not match
  if $facts['win_collector_path'] != $collector_path or $facts['win_collector_version'] != $version {
    file { "${facts['win_temp']}\\${msi_name}":
      source => "${repo_url}/${msi_name}",
    }

    -> package { $package_name:
      ensure => $version,
      source => "${facts['win_temp']}\\${msi_name}",
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
