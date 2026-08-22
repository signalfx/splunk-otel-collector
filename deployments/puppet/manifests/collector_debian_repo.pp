# Installs the collector debian package repository config
class splunk_otel_collector::collector_debian_repo ($repo_url, $package_stage, $repo, $apt_gpg_key, $manage_repo) {
  if $manage_repo {
    $apt_keyring_path = '/etc/apt/keyrings/splunk-otel-collector.gpg'

    file { '/etc/apt/keyrings':
      ensure => directory,
      owner  => 'root',
      group  => 'root',
      mode   => '0755',
    }
    -> file { $apt_keyring_path:
      ensure => file,
      source => $apt_gpg_key,
      owner  => 'root',
      group  => 'root',
      mode   => '0644',
    }
    -> apt::source { 'splunk-otel-collector':
      location => $repo_url,
      release  => $package_stage,
      repos    => $repo,
      keyring  => $apt_keyring_path,
    }
  } else {
    file { '/etc/apt/sources.list.d/splunk-otel-collector.list':
      ensure => absent,
    }
  }
}
