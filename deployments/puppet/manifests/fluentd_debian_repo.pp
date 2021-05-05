# Installs the fluentd debian package repository config
class splunk_otel_connector::fluentd_debian_repo ($repo_url, $gpg_key_url, $version, $manage_repo) {
  $distro = downcase($facts['os']['distro']['id'])
  $codename = downcase($facts['os']['distro']['codename'])
  $major_version = $version.split('\.')[0]

  if $manage_repo {
    if $distro != 'ubuntu' and $distro != 'debian' {
      fail("Your distribution '${distro}' is not currently supported")
    }

    apt::source { 'splunk-td-agent':
      location => "${repo_url}/${major_version}/${distro}/${codename}",
      release  => $codename,
      repos    => 'contrib',
      key      => {
        id     => 'BEE682289B2217F45AF4CC3F901F9177AB97ACBE',
        source => $gpg_key_url,
      },
    }
  } else {
    file { '/etc/apt/sources.list.d/splunk-td-agent.list':
      ensure => 'absent',
    }
  }
}
