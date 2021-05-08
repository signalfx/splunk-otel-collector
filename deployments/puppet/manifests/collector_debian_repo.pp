# Installs the collector debian package repository config
class splunk_otel_connector::collector_debian_repo ($repo_url, $package_stage, $repo, $apt_gpg_key, $manage_repo) {

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
