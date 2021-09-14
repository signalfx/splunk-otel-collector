# Installs the collector yum package repostitory for the given stage
class splunk_otel_collector::collector_yum_repo ($repo_url, $yum_gpg_key, $manage_repo, $repo_path) {

  if $manage_repo {
    file { "${repo_path}/splunk-otel-collector.repo":
      content => @("EOH")
        [splunk-otel-collector]
        name=Splunk OpenTelemetry Collector
        baseurl=${repo_url}
        gpgcheck=1
        gpgkey=${yum_gpg_key}
        enabled=1
        | EOH
      ,
      mode    => '0644',
    }
  } else {
    file { "${repo_path}/splunk-otel-collector.repo":
      ensure => absent,
    }
  }
}
