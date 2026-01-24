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
  }

  $collector_service_name = 'splunk-otel-collector'
  $collector_package_name = $facts['os']['family'] ? {
    'windows' => 'Splunk OpenTelemetry Collector',
    default   => $collector_service_name,
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
