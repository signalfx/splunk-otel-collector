# Main class that installs and configures the agent
class splunk_otel_collector (
  $splunk_access_token     = '',  # required
  $splunk_realm            = '',  # required
  $splunk_ingest_url       = "https://ingest.${splunk_realm}.signalfx.com",
  $splunk_api_url          = "https://api.${splunk_realm}.signalfx.com",
  $splunk_trace_url        = "${splunk_ingest_url}/v2/trace",
  $splunk_hec_url          = "${splunk_ingest_url}/v1/log",
  $splunk_hec_token        = $splunk_access_token,
  $splunk_bundle_dir       = $splunk_otel_collector::params::splunk_bundle_dir,
  $splunk_collectd_dir     = $splunk_otel_collector::params::splunk_collectd_dir,
  $splunk_memory_total_mib = '512',
  $splunk_ballast_size_mib = '',
  $collector_version       = $splunk_otel_collector::params::collector_version,
  $collector_config_source = $splunk_otel_collector::params::collector_config_source,
  $collector_config_dest   = $splunk_otel_collector::params::collector_config_dest,
  $package_stage           = 'release',  # collector package repository stage: release, beta, or test
  $apt_repo_url            = 'https://splunk.jfrog.io/splunk/otel-collector-deb',
  $apt_repo                = 'main',
  $yum_repo_url            = "https://splunk.jfrog.io/splunk/otel-collector-rpm/${package_stage}/\$basearch",
  $win_repo_url            = "https://dl.signalfx.com/splunk-otel-collector/msi/${package_stage}",
  $service_user            = 'splunk-otel-collector',  # linux only
  $service_group           = 'splunk-otel-collector',  # linux only
  $apt_gpg_key             = 'https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg',
  $yum_gpg_key             = 'https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub',
  $with_fluentd            = true,
  $fluentd_repo_base       = 'https://packages.treasuredata.com',
  $fluentd_gpg_key         = 'https://packages.treasuredata.com/GPG-KEY-td-agent',
  $fluentd_version         = $splunk_otel_collector::params::fluentd_version,
  $fluentd_config_source   = $splunk_otel_collector::params::fluentd_config_source,
  $fluentd_config_dest     = $splunk_otel_collector::params::fluentd_config_dest,
  $fluentd_capng_c_version = '<=0.2.2',  # linux only
  $fluentd_systemd_version = '<=1.0.2',  # linux only
  $manage_repo             = true  # linux only
) inherits splunk_otel_collector::params {

  $collector_service_name = 'splunk-otel-collector'
  $collector_package_name = $::osfamily ? {
    'windows' => 'Splunk OpenTelemetry Collector',
    default   => $collector_service_name,
  }
  $fluentd_service_name = $::osfamily ? {
    'windows' => 'fluentdwinsvc',
    default   => 'td-agent',
  }
  $fluentd_package_name = $::osfamily ? {
    'windows' => "Td-agent v${fluentd_version}",
    default   => $fluentd_service_name,
  }

  if empty($splunk_access_token) {
    fail('The splunk_access_token parameter is required')
  }

  if empty($splunk_realm) {
    fail('The splunk_realm parameter is required')
  }

  if $::osfamily == 'windows' {
    if empty($collector_version) {
      fail('The $collector_version parameter is required for Windows')
    }
    if $with_fluentd and empty($fluentd_version) {
      fail('The $fluentd_version parameter is required for Windows')
    }
  }

  case $::osfamily {
    'debian': {
      if $facts['service_provider'] != 'systemd' {
        fail('Only systemd is currently supported')
      }
      class { 'splunk_otel_collector::collector_debian_repo':
        repo_url      => $apt_repo_url,
        package_stage => $package_stage,
        repo          => $apt_repo,
        apt_gpg_key   => $apt_gpg_key,
        manage_repo   => $manage_repo,
      }
      -> package { $collector_package_name:
        ensure  => $collector_version,
        require => Exec['apt_update'],
      }
    }
    'redhat': {
      if $facts['service_provider'] != 'systemd' {
        fail('Only systemd is currently supported')
      }
      package { 'libcap':
        ensure => 'installed',
      }
      class { 'splunk_otel_collector::collector_yum_repo':
        repo_url    => $yum_repo_url,
        yum_gpg_key => $yum_gpg_key,
        manage_repo => $manage_repo,
        repo_path   => '/etc/yum.repos.d',
      }
      -> package { $collector_package_name:
        ensure  => $collector_version,
        require => Package['libcap'],
      }
    }
    'suse': {
      if $facts['service_provider'] != 'systemd' {
        fail('Only systemd is currently supported')
      }
      package { 'libcap-progs':
        ensure => 'installed',
      }
      # Workaround for older zypper versions that have issues importing gpg keys
      exec { 'Import yum gpg key':
        command => "rpm --import ${yum_gpg_key}",
        path    => ['/bin', '/sbin', '/usr/bin', '/usr/sbin'],
      }
      -> class { 'splunk_otel_collector::collector_yum_repo':
        repo_url    => $yum_repo_url,
        yum_gpg_key => $yum_gpg_key,
        manage_repo => $manage_repo,
        repo_path   => '/etc/zypp/repos.d',
      }
      -> package { $collector_package_name:
        ensure  => $collector_version,
        require => Package['libcap-progs'],
      }
    }
    'windows': {
      class { 'splunk_otel_collector::collector_win_install':
        repo_url     => $win_repo_url,
        version      => $collector_version,
        package_name => $collector_package_name,
        service_name => $collector_service_name,
      }
    }
    default: {
      fail("Your OS (${::osfamily}) is not supported by the Splunk OpenTelemetry Collector")
    }
  }

  if $::osfamily != 'windows' {
    $collector_config_dir = $collector_config_dest.split('/')[0, - 2].join('/')
    $env_file_path = '/etc/otel/collector/splunk-otel-collector.conf'
    $env_file_content = @("EOH")
      SPLUNK_ACCESS_TOKEN=${splunk_access_token}
      SPLUNK_API_URL=${splunk_api_url}
      SPLUNK_BALLAST_SIZE_MIB=${splunk_ballast_size_mib}
      SPLUNK_BUNDLE_DIR=${splunk_bundle_dir}
      SPLUNK_COLLECTD_DIR=${splunk_collectd_dir}
      SPLUNK_CONFIG=${collector_config_dest}
      SPLUNK_HEC_TOKEN=${splunk_hec_token}
      SPLUNK_HEC_URL=${splunk_hec_url}
      SPLUNK_INGEST_URL=${splunk_ingest_url}
      SPLUNK_MEMORY_TOTAL_MIB=${splunk_memory_total_mib}
      SPLUNK_REALM=${splunk_realm}
      SPLUNK_TRACE_URL=${splunk_trace_url}
      | EOH

    class { 'splunk_otel_collector::collector_service_owner':
      service_name  => $collector_service_name,
      service_user  => $service_user,
      service_group => $service_group,
    }

    -> file { $env_file_path:
      content => $env_file_content,
      mode    => '0600',
      owner   => $service_user,
      group   => $service_group,
    }

    exec { 'create collector config directory':
      command => "mkdir -p ${collector_config_dir}",
      path    => ['/bin', '/sbin', '/usr/bin', '/usr/sbin'],
      unless  => "test -d ${collector_config_dir}",
    }

    -> file { $collector_config_dest:
      source  => $collector_config_source,
      require => Package[$collector_package_name],
    }

    -> service { $collector_service_name:
      ensure    => true,
      enable    => true,
      subscribe => [File[$collector_config_dest, $env_file_path]],
    }
  } else {
    file { $collector_config_dest:
      source  => $collector_config_source,
      require => Class['splunk_otel_collector::collector_win_install'],
    }

    -> class { 'splunk_otel_collector::collector_win_registry': }

    -> service { $collector_service_name:
      ensure    => true,
      enable    => true,
      subscribe => [Class['splunk_otel_collector::collector_win_registry'], File[$collector_config_dest]],
    }
  }

  if $with_fluentd and $::osfamily != 'suse' {
    case $::osfamily {
      'debian': {
        package { ['build-essential', 'libcap-ng0', 'libcap-ng-dev', 'pkg-config']:
          ensure  => 'installed',
          require => Exec['apt_update'],
        }
        class { 'splunk_otel_collector::fluentd_debian_repo':
          repo_url    => $fluentd_repo_base,
          gpg_key_url => $fluentd_gpg_key,
          version     => $fluentd_version,
          manage_repo => $manage_repo,
        }
        -> package { $fluentd_package_name:
          ensure  => $fluentd_version,
          require => Exec['apt_update'],
        }
        package { 'capng_c':
          ensure   => $fluentd_capng_c_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => Package[$fluentd_package_name, 'build-essential', 'libcap-ng0', 'libcap-ng-dev', 'pkg-config'],
        }
        package { 'fluent-plugin-systemd':
          ensure   => $fluentd_systemd_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => Package[$fluentd_package_name, 'build-essential', 'libcap-ng0', 'libcap-ng-dev', 'pkg-config'],
        }
      }
      'redhat': {
        class { 'splunk_otel_collector::fluentd_yum_repo':
          repo_url    => $fluentd_repo_base,
          gpg_key_url => $fluentd_gpg_key,
          version     => $fluentd_version,
          manage_repo => $manage_repo,
        }
        -> package { $fluentd_package_name:
          ensure => $fluentd_version,
        }
        yum::group { 'Development Tools':
          ensure => 'present',
        }
        package { ['libcap-ng', 'libcap-ng-devel', 'pkgconfig']:
          ensure => 'installed',
        }
        package { 'capng_c':
          ensure   => $fluentd_capng_c_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => [Package[$fluentd_package_name, 'libcap-ng', 'libcap-ng-devel', 'pkgconfig'],
            Yum::Group['Development Tools']],
        }
        package { 'fluent-plugin-systemd':
          ensure   => $fluentd_systemd_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => [Package[$fluentd_package_name, 'libcap-ng', 'libcap-ng-devel', 'pkgconfig'],
            Yum::Group['Development Tools']],
        }
      }
      'windows': {
        class { 'splunk_otel_collector::fluentd_win_install':
          repo_base    => $fluentd_repo_base,
          version      => $fluentd_version,
          package_name => $fluentd_package_name,
          service_name => $fluentd_service_name,
        }
      }
      default: {
        fail("Your OS (${::osfamily}) is not supported by the Splunk OpenTelemetry Collector")
      }
    }

    if $::osfamily != 'windows' {
      $fluentd_config_dir = $fluentd_config_dest.split('/')[0, - 2].join('/')
      $fluentd_config_override = '/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf'
      $fluentd_config_override_dir= $fluentd_config_override.split('/')[0, - 2].join('/')

      exec { 'create fluentd config directory':
        command => "mkdir -p ${fluentd_config_dir}",
        path    => ['/bin', '/sbin', '/usr/bin', '/usr/sbin'],
        unless  => "test -d ${fluentd_config_dir}",
      }

      -> file { $fluentd_config_dest:
        source  => $fluentd_config_source,
        require => Package[$collector_package_name],
      }

      file { $fluentd_config_override_dir:
        ensure => 'directory',
      }

      -> file { $fluentd_config_override:
        content => @("EOH")
          [Service]
          Environment=FLUENT_CONF=${fluentd_config_dest}
          | EOH
        ,
        require => File[$fluentd_config_dest],
      }

      # enable linux capabilities for fluentd
      exec { 'fluent-cap-ctl':
        command => '/opt/td-agent/bin/fluent-cap-ctl --add "dac_override,dac_read_search" -f /opt/td-agent/bin/ruby',
        path    => ['/bin', '/sbin', '/usr/bin', '/usr/sbin'],
        require => Package['capng_c'],
        onlyif  => 'test -f /opt/td-agent/bin/fluent-cap-ctl',
      }

      -> service { $fluentd_service_name:
        ensure    => true,
        enable    => true,
        require   => [Package[$fluentd_package_name, 'fluent-plugin-systemd'], Service[$collector_service_name]],
        subscribe => File[$fluentd_config_dest, $fluentd_config_override],
      }
    } else {
      $collector_install_dir = "${::win_programfiles}\\Splunk\\OpenTelemetry Collector"
      $td_agent_config_dir = "${::win_systemdrive}\\opt\\td-agent\\etc\\td-agent"
      $td_agent_config_dest = "${td_agent_config_dir}\\td-agent.conf"

      file { $td_agent_config_dest:
        source  => $fluentd_config_source,
        require => Class['splunk_otel_collector::collector_win_install', 'splunk_otel_collector::fluentd_win_install'],
      }

      file { "${td_agent_config_dir}\\conf.d":
        ensure  => 'directory',
        source  => "${collector_install_dir}\\fluentd\\conf.d",
        recurse => true,
        require => Class['splunk_otel_collector::collector_win_install', 'splunk_otel_collector::fluentd_win_install'],
      }

      exec { "Stop ${fluentd_service_name}":
        command     => "Stop-Service -Name \'${fluentd_service_name}\'",
        # lint:ignore:140chars
        onlyif      => "((Get-CimInstance -ClassName win32_service -Filter 'Name = \'${fluentd_service_name}\'' | Select Name, State).Name)",
        # lint:endignore
        provider    => 'powershell',
        subscribe   => [
          Class['splunk_otel_collector::fluentd_win_install'],
          File[$td_agent_config_dest, "${td_agent_config_dir}\\conf.d"]
        ],
        refreshonly => true,
      }

      ~> service { $fluentd_service_name:
        ensure  => true,
        enable  => true,
        require => [Class['splunk_otel_collector::fluentd_win_install'], Service[$collector_service_name]],
      }
    }
  }
}
