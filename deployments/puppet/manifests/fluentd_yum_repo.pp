# Installs the fluentd yum package repostitory
#
# @param repo_url
# @param gpg_key_url
# @param version
# @param manage_repo
#
class splunk_otel_collector::fluentd_yum_repo (
  String $repo_url,
  String $gpg_key_url,
  String $version,
  Boolean $manage_repo,
) {
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
      mode    => '0644',
    }
  } else {
    file { '/etc/yum.repos.d/splunk-td-agent.repo':
      ensure => 'absent',
    }
  }
}
