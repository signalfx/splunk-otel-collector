# Download and install the fluentd MSI on Windows
#
# @param repo_base
# @param version
# @param package_name
# @param service_name
#
class splunk_otel_collector::fluentd_win_install (
  String $repo_base,
  String $version,
  String $package_name,
  String $service_name,
) {
  $msi_name = "td-agent-${version}-x64.msi"

  file { "${facts['win_temp']}\\${msi_name}":
    source => "${repo_base}/4/windows/${msi_name}",
  }

  -> package { $package_name:
    ensure => $version,
    source => "${facts['win_temp']}\\${msi_name}",
  }
}
