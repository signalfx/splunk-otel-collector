# Installs the fluentd yum package repostitory
class splunk_otel_connector::fluentd_yum_repo ($repo_url, $gpg_key_url, $version, $manage_repo) {

  if $manage_repo {
    $os_name = $facts['os']['name'] ? {
      'Amazon' => 'amazon',
      default  => 'redhat',
    }
    $major_version = $version.split('\.')[0]
    $url = "${repo_url}/${major_version}/${os_name}/\$releasever/\$basearch"

    file { '/etc/yum.repos.d/splunk-td-agent.repo':
      content => @("EOH")
        [td-agent]
        name=TreasureData Repository
        baseurl=${url}
        gpgcheck=1
        gpgkey=${gpg_key_url}
        enabled=1
        | EOH
    ,
    mode      => '0644',
    }
  } else {
    file { '/etc/yum.repos.d/splunk-td-agent.repo':
      ensure => 'absent',
    }
  }
}
