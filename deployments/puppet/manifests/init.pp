# Main class that installs and configures the agent
class splunk_otel_collector (
  $splunk_access_token     = '',  # required
  $splunk_realm            = '',  # required
  $splunk_ingest_url       = "https://ingest.${splunk_realm}.signalfx.com",
  $splunk_api_url          = "https://api.${splunk_realm}.signalfx.com",
  $splunk_hec_url          = "${splunk_ingest_url}/v1/log",
  $splunk_hec_token        = $splunk_access_token,
  $splunk_bundle_dir       = $splunk_otel_collector::params::splunk_bundle_dir,
  $splunk_collectd_dir     = $splunk_otel_collector::params::splunk_collectd_dir,
  $splunk_memory_total_mib = '512',
  $splunk_listen_interface = '',
  $collector_command_line_args = '',
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
  $with_fluentd            = false,
  $fluentd_repo_base       = $splunk_otel_collector::params::fluentd_base_url,
  $fluentd_gpg_key         = 'https://packages.treasuredata.com/GPG-KEY-td-agent',
  $fluentd_version         = $splunk_otel_collector::params::fluentd_version,
  $fluentd_config_source   = $splunk_otel_collector::params::fluentd_config_source,
  $fluentd_config_dest     = $splunk_otel_collector::params::fluentd_config_dest,
  $fluentd_capng_c_version = '<=0.2.2',  # linux only
  $fluentd_systemd_version = '<=1.0.2',  # linux only
  $gomemlimit              = '',
  $manage_repo             = true,  # linux only
  $with_auto_instrumentation                = false,  # linux only
  $auto_instrumentation_version             = $splunk_otel_collector::params::auto_instrumentation_version,  # linux only
  $auto_instrumentation_systemd             = false,  # linux only
  $auto_instrumentation_ld_so_preload       = '',  # linux only
  $auto_instrumentation_java_agent_jar      = $splunk_otel_collector::params::auto_instrumentation_java_agent_jar,  # linux only
  $auto_instrumentation_resource_attributes = '',  # linux only
  $auto_instrumentation_service_name        = '',  # linux only
  $auto_instrumentation_generate_service_name   = true,   # linux only
  $auto_instrumentation_disable_telemetry       = false,  # linux only
  $auto_instrumentation_enable_profiler         = false,  # linux only
  $auto_instrumentation_enable_profiler_memory  = false,  # linux only
  $auto_instrumentation_enable_metrics          = false,  # linux only
  $auto_instrumentation_otlp_endpoint           = '',  # linux only
  $auto_instrumentation_otlp_endpoint_protocol  = '',  # linux only
  $auto_instrumentation_metrics_exporter        = '',  # linux only
  $auto_instrumentation_logs_exporter           = '',  # linux only
  $with_auto_instrumentation_sdks               = ['java', 'nodejs', 'dotnet'], # linux only
  $auto_instrumentation_npm_path                = 'npm', # linux only
  $collector_additional_env_vars            = {}
) inherits splunk_otel_collector::params {
  if empty($splunk_access_token) {
    fail('The splunk_access_token parameter is required')
  }

  if empty($splunk_realm) {
    fail('The splunk_realm parameter is required')
  }

  if $facts['os']['family'] == 'windows' {
    if empty($collector_version) {
      fail('The $collector_version parameter is required for Windows')
    }
    if $with_fluentd and empty($fluentd_version) {
      fail('The $fluentd_version parameter is required for Windows')
    }
  }

  $collector_service_name = 'splunk-otel-collector'
  $collector_package_name = $facts['os']['family'] ? {
    'windows' => 'Splunk OpenTelemetry Collector',
    default   => $collector_service_name,
  }
  $fluentd_service_name = $facts['os']['family'] ? {
    'windows' => 'fluentdwinsvc',
    default   => 'td-agent',
  }
  $fluentd_package_name = $facts['os']['family'] ? {
    'windows' => "Td-agent v${fluentd_version}",
    default   => $fluentd_service_name,
  }

  if $facts['os']['family'] == 'suse' or ('amazon' in downcase($facts['os']['name']) and $facts['os']['release']['major'] == '2023') {
    $install_fluentd = false
  } else {
    $install_fluentd = $with_fluentd
  }

  case $facts['os']['family'] {
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
      fail("Your OS (${facts['os']['family']}) is not supported by the Splunk OpenTelemetry Collector")
    }
  }

  if $facts['os']['family'] != 'windows' {
    $collector_config_dir = $collector_config_dest.split('/')[0, - 2].join('/')
    $env_file_path = '/etc/otel/collector/splunk-otel-collector.conf'

    class { 'splunk_otel_collector::collector_service_owner':
      service_name  => $collector_service_name,
      service_user  => $service_user,
      service_group => $service_group,
    }

    -> file { $env_file_path:
      ensure  => file,
      content => template('splunk_otel_collector/splunk-otel-collector.conf.erb'),
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
    if $collector_config_source != $splunk_otel_collector::params::default_win_config_file {
      file { $collector_config_dest:
        source  => $collector_config_source,
        require => Class['splunk_otel_collector::collector_win_install'],
      }
    } else {
      file { $collector_config_dest:
        ensure  => file,
        require => Class['splunk_otel_collector::collector_win_install'],
      }
    }

    -> class { 'splunk_otel_collector::collector_win_registry': }

    -> service { $collector_service_name:
      ensure    => true,
      enable    => true,
      subscribe => [Class['splunk_otel_collector::collector_win_registry'], File[$collector_config_dest]],
    }
  }

  if $install_fluentd {
    deprecation('with_fluentd', 'Fluentd support has been deprecated and will be removed in a future release. Please refer to documentation on how to replace usage: https://github.com/signalfx/splunk-otel-collector/blob/main/docs/deprecations/fluentd-support.md')

    case $facts['os']['family'] {
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
        fail("Your OS (${facts['os']['family']}) is not supported by the Splunk OpenTelemetry Collector")
      }
    }

    if $facts['os']['family'] != 'windows' {
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
        notify  => Exec['systemctl daemon-reload'],
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
      $collector_install_dir = "${facts['win_programfiles']}\\Splunk\\OpenTelemetry Collector"
      $td_agent_config_dir = "${facts['win_systemdrive']}\\opt\\td-agent\\etc\\td-agent"
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

  if $with_auto_instrumentation {
    $auto_instrumentation_package_name = 'splunk-otel-auto-instrumentation'
    $ld_so_preload_path =  '/etc/ld.so.preload'
    $libsplunk_path = '/usr/lib/splunk-instrumentation/libsplunk.so'
    $instrumentation_config_path = '/usr/lib/splunk-instrumentation/instrumentation.conf'
    $zeroconfig_java_config_path = '/etc/splunk/zeroconfig/java.conf'
    $zeroconfig_node_config_path = '/etc/splunk/zeroconfig/node.conf'
    $zeroconfig_dotnet_config_path = '/etc/splunk/zeroconfig/dotnet.conf'
    $zeroconfig_systemd_config_path = '/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf'
    $with_new_instrumentation = $auto_instrumentation_version == 'latest' or versioncmp($auto_instrumentation_version, '0.87.0') >= 0
    $dotnet_supported = $facts['os']['architecture'] in ['amd64', 'x86_64'] and ($auto_instrumentation_version == 'latest' or versioncmp($auto_instrumentation_version, '0.99.0') >= 0) # lint:ignore:140chars

    if $facts['os']['family'] == 'debian' {
      package { $auto_instrumentation_package_name:
        ensure  => $auto_instrumentation_version,
        require => [Class['splunk_otel_collector::collector_debian_repo'], Package[$collector_package_name]],
      }
    } elsif $facts['os']['family'] == 'redhat' or $facts['os']['family'] == 'suse' {
      package { $auto_instrumentation_package_name:
        ensure  => $auto_instrumentation_version,
        require => [Class['splunk_otel_collector::collector_yum_repo'], Package[$collector_package_name]],
      }
    } else {
      fail("Splunk OpenTelemetry Auto Instrumentation is not supported for your OS family (${facts['os']['family']})")
    }

    if 'nodejs' in $with_auto_instrumentation_sdks and $with_new_instrumentation {
      $splunk_otel_js_path = '/usr/lib/splunk-instrumentation/splunk-otel-js.tgz'
      $splunk_otel_js_prefix = '/usr/lib/splunk-instrumentation/splunk-otel-js'

      file { [$splunk_otel_js_prefix, "${$splunk_otel_js_prefix}/node_modules"]:
        ensure  => 'directory',
        require => Package[$auto_instrumentation_package_name],
      }
      -> exec { 'Install splunk-otel-js':
        command  => "${$auto_instrumentation_npm_path} install ${$splunk_otel_js_path}",
        provider => shell,
        cwd      => $splunk_otel_js_prefix,
        require  => Package[$auto_instrumentation_package_name],
      }
    }

    file { $ld_so_preload_path:
      ensure  => file,
      content => template('splunk_otel_collector/ld.so.preload.erb'),
      require => Package[$auto_instrumentation_package_name],
    }

    if $auto_instrumentation_systemd {
      file { [$zeroconfig_java_config_path, $zeroconfig_node_config_path, $zeroconfig_dotnet_config_path, $instrumentation_config_path]:
        ensure  => absent,
        require => Package[$auto_instrumentation_package_name],
      }
      file { ['/usr/lib/systemd', '/usr/lib/systemd/system.conf.d']:
        ensure => directory,
      }
      -> file { $zeroconfig_systemd_config_path:
        ensure  => file,
        content => template('splunk_otel_collector/00-splunk-otel-auto-instrumentation.conf.erb'),
        require => Package[$auto_instrumentation_package_name],
        notify  => Exec['systemctl daemon-reload'],
      }
    } else {
      file { $zeroconfig_systemd_config_path:
        ensure => absent,
      }
      if $with_new_instrumentation {
        if 'java' in $with_auto_instrumentation_sdks {
          file { $zeroconfig_java_config_path:
            ensure  => file,
            content => template('splunk_otel_collector/java.conf.erb'),
            require => Package[$auto_instrumentation_package_name],
          }
        } else {
          file { $zeroconfig_java_config_path:
            ensure => absent,
          }
        }
        if 'nodejs' in $with_auto_instrumentation_sdks {
          file { $zeroconfig_node_config_path:
            ensure  => file,
            content => template('splunk_otel_collector/node.conf.erb'),
            require => Exec['Install splunk-otel-js'],
          }
        } else {
          file { $zeroconfig_node_config_path:
            ensure => absent,
          }
        }
        if 'dotnet' in $with_auto_instrumentation_sdks and $dotnet_supported {
          file { $zeroconfig_dotnet_config_path:
            ensure  => file,
            content => template('splunk_otel_collector/dotnet.conf.erb'),
            require => Package[$auto_instrumentation_package_name],
          }
        } else {
          file { $zeroconfig_dotnet_config_path:
            ensure => absent,
          }
        }
      } else {
        file { $instrumentation_config_path:
          ensure  => file,
          content => template('splunk_otel_collector/instrumentation.conf.erb'),
          require => Package[$auto_instrumentation_package_name],
        }
      }
    }
  }
}
