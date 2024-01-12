# Installs the collector debian package repository config
#
# @param repo_url
# @param package_stage
# @param repo
# @param apt_gpg_key
# @param manage_repo
#
class splunk_otel_collector::collector_debian_repo (
  String $repo_url,
  String $package_stage,
  String $repo,
  String $apt_gpg_key,
  Boolean $manage_repo,
) {
  if $manage_repo {
    apt::source { 'splunk-otel-collector':
      location => $repo_url,
      release  => $package_stage,
      repos    => $repo,
      key      => {
        id     => '58C33310B7A354C1279DB6695EFA01EDB3CD4420',
        source => $apt_gpg_key,
      },
    }
  } else {
    file { '/etc/apt/sources.list.d/splunk-otel-collector.list':
      ensure => absent,
    }
  }
}
