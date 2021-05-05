# Main class that installs and configures the agent
class splunk_otel_connector (
  $splunk_access_token     = '',  # required
  $collector_config_source = 'file:///etc/otel/collector/agent_config.yaml',
  $collector_config_dest   = '/etc/otel/collector/agent_config.yaml',
  $splunk_realm            = 'us0',
  $splunk_ingest_url       = "https://ingest.${splunk_realm}.signalfx.com",
  $splunk_api_url          = "https://api.${splunk_realm}.signalfx.com",
  $splunk_trace_url        = "${splunk_ingest_url}/v2/trace",
  $splunk_hec_url          = "${splunk_ingest_url}/v1/log",
  $splunk_hec_token        = $splunk_access_token,
  $splunk_bundle_dir       = '/usr/lib/splunk-otel-collector/agent-bundle',
  $splunk_collectd_dir     = "${splunk_bundle_dir}/run/collectd",
  $package_stage           = 'release',  # collector package repository stage: release, beta, or test
  $apt_repo_url            = 'https://splunk.jfrog.io/splunk/otel-collector-deb',
  $apt_repo                = 'main',
  $yum_repo_url            = "https://splunk.jfrog.io/splunk/otel-collector-rpm/${package_stage}/\$basearch",
  $collector_version       = 'latest',
  $service_user            = 'splunk-otel-collector',  # linux only
  $service_group           = 'splunk-otel-collector',  # linux only
  $apt_gpg_key             = 'https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg',
  $yum_gpg_key             = 'https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub',
  $with_fluentd            = true,
  $fluentd_repo_base       = 'https://packages.treasuredata.com',
  $fluentd_gpg_key         = 'https://packages.treasuredata.com/GPG-KEY-td-agent',
  $fluentd_version         = '4.1.0',
  $fluentd_version_jessie  = '3.3.0-1',
  $fluentd_version_stretch = '3.7.1-0',
  $fluentd_config_source   = 'file:///etc/otel/collector/fluentd/fluent.conf',
  $fluentd_config_dest     = '/etc/otel/collector/fluentd/fluent.conf',
  $fluentd_capng_c_version = '<=0.2.2',
  $fluentd_systemd_version = '<=1.0.2',
  $manage_repo             = true  # linux only
) {

  if empty($splunk_access_token) {
    fail('The splunk_access_token parameter is required')
  }

  if $::osfamily == 'windows' {
    fail('Windows is not currently supported')
  } else {
    if $facts['service_provider'] != 'systemd' {
      fail('Only systemd is currently supported')
    }
    $collector_config_dir = $collector_config_dest.split('/')[0, - 2].join('/')
  }

  $collector_service_name = 'splunk-otel-collector'

  case $::osfamily {
    'debian': {
      class { 'splunk_otel_connector::collector_debian_repo':
        repo_url      => $apt_repo_url,
        package_stage => $package_stage,
        repo          => $apt_repo,
        apt_gpg_key   => $apt_gpg_key,
        manage_repo   => $manage_repo,
      }
      -> package { $collector_service_name:
        ensure  => $collector_version,
        require => Exec['apt_update'],
      }
    }
    'redhat': {
      class { 'splunk_otel_connector::collector_yum_repo':
        repo_url    => $yum_repo_url,
        yum_gpg_key => $yum_gpg_key,
        manage_repo => $manage_repo,
      }
      -> package { $collector_service_name:
        ensure => $collector_version,
      }
    }
    default: {
      fail("Your OS (${::osfamily}) is not supported by the Splunk OpenTelemetry Connector")
    }
  }

  if $::osfamily != 'windows' {
    $env_file_path = '/etc/otel/collector/splunk-otel-collector.conf'
    $env_file_content = @("EOH")
      SPLUNK_CONFIG=${collector_config_dest}
      SPLUNK_ACCESS_TOKEN=${splunk_access_token}
      SPLUNK_API_URL=${splunk_api_url}
      SPLUNK_BUNDLE_DIR=${splunk_bundle_dir}
      SPLUNK_COLLECTD_DIR=${splunk_collectd_dir}
      SPLUNK_HEC_TOKEN=${splunk_hec_token}
      SPLUNK_HEC_URL=${splunk_hec_url}
      SPLUNK_INGEST_URL=${splunk_ingest_url}
      SPLUNK_REALM=${splunk_realm}
      SPLUNK_TRACE_URL=${splunk_trace_url}
      | EOH

    class { 'splunk_otel_connector::collector_service_owner':
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
      require => Package[$collector_service_name],
    }

    service { $collector_service_name:
      ensure    => true,
      enable    => true,
      subscribe => [Package[$collector_service_name], File[$collector_config_dest, $env_file_path]],
    }
  }

  if $with_fluentd {
    $fluentd_service_name = 'td-agent'
    $fluentd_config_dir = $fluentd_config_dest.split('/')[0, - 2].join('/')
    $fluentd_config_override = '/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf'
    $fluentd_config_override_dir= $fluentd_config_override.split('/')[0, - 2].join('/')

    case $::osfamily {
      'debian': {
        package { ['build-essential', 'libcap-ng0', 'libcap-ng-dev', 'pkg-config']:
          ensure  => 'installed',
          require => Exec['apt_update'],
        }
        case downcase($facts['os']['distro']['codename']) {
          'jessie': {
            $version = $fluentd_version_jessie
          }
          'stretch': {
            $version = $fluentd_version_stretch
          }
          default: {
            $version = "${fluentd_version}-1"
          }
        }
        class { 'splunk_otel_connector::fluentd_debian_repo':
          repo_url    => $fluentd_repo_base,
          gpg_key_url => $fluentd_gpg_key,
          version     => $version,
          manage_repo => $manage_repo,
        }
        -> package { $fluentd_service_name:
          ensure  => $version,
          require => Exec['apt_update'],
        }
        package { 'capng_c':
          ensure   => $fluentd_capng_c_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => Package[$fluentd_service_name, 'build-essential', 'libcap-ng0', 'libcap-ng-dev', 'pkg-config'],
        }
        package { 'fluent-plugin-systemd':
          ensure   => $fluentd_systemd_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => Package[$fluentd_service_name, 'build-essential', 'libcap-ng0', 'libcap-ng-dev', 'pkg-config'],
        }
      }
      'redhat': {
        class { 'splunk_otel_connector::fluentd_yum_repo':
          repo_url    => $fluentd_repo_base,
          gpg_key_url => $fluentd_gpg_key,
          version     => $fluentd_version,
          manage_repo => $manage_repo,
        }
        -> package { $fluentd_service_name:
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
          require  => [Package[$fluentd_service_name, 'libcap-ng', 'libcap-ng-devel', 'pkgconfig'],
            Yum::Group['Development Tools']],
        }
        package { 'fluent-plugin-systemd':
          ensure   => $fluentd_systemd_version,
          provider => gem,
          command  => '/usr/sbin/td-agent-gem',
          require  => [Package[$fluentd_service_name, 'libcap-ng', 'libcap-ng-devel', 'pkgconfig'],
            Yum::Group['Development Tools']],
        }
      }
      default: {
        fail("Your OS (${::osfamily}) is not supported by the Splunk OpenTelemetry Connector")
      }
    }

    exec { 'create fluentd config directory':
      command => "mkdir -p ${fluentd_config_dir}",
      path    => ['/bin', '/sbin', '/usr/bin', '/usr/sbin'],
      unless  => "test -d ${fluentd_config_dir}",
    }

    -> file { $fluentd_config_dest:
      source  => $fluentd_config_source,
      require => Package[$collector_service_name],
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
      require   => [Package[$fluentd_service_name, 'fluent-plugin-systemd'], Service[$collector_service_name]],
      subscribe => File[$fluentd_config_dest, $fluentd_config_override],
    }
  }
}
