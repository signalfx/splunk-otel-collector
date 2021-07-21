# Download and install the fluentd MSI on Windows
class splunk_otel_collector::fluentd_win_install ($repo_base, $version, $package_name, $service_name) {
  $msi_name = "td-agent-${version}-x64.msi"

  file { "${::win_temp}\\${msi_name}":
    source => "${repo_base}/4/windows/${msi_name}"
  }

  -> package { $package_name:
    ensure => $version,
    source => "${::win_temp}\\${msi_name}",
  }
}
