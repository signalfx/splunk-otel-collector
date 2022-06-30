# Class for default param values based on OS
class splunk_otel_collector::params {
  $fluentd_version_default = '4.3.2'
  $fluentd_version_stretch = '3.7.1-0'

  if $::osfamily == 'redhat' or $::osfamily == 'debian' or $::osfamily == 'suse' {
    $collector_version = 'latest'
    $collector_config_dir = '/etc/otel/collector'
    $fluentd_config_dir = "${collector_config_dir}/fluentd"
    $splunk_bundle_dir = '/usr/lib/splunk-otel-collector/agent-bundle'
    $splunk_collectd_dir = "${splunk_bundle_dir}/run/collectd"
    $collector_config_source = "${collector_config_dir}/agent_config.yaml"
    $collector_config_dest = $collector_config_source
    if $::osfamily == 'debian' {
      $fluentd_version = downcase($facts['os']['distro']['codename']) ? {
        'stretch' => $fluentd_version_stretch,
        default   => "${fluentd_version_default}-1",
      }
    } else {
      $fluentd_version = $fluentd_version_default
    }
    $fluentd_config_source = "${fluentd_config_dir}/fluent.conf"
    $fluentd_config_dest = $fluentd_config_source
    $auto_instrumentation_version = 'latest'
    $auto_instrumentation_java_agent_jar = '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar'

  } elsif $::osfamily == 'windows' {
    $collector_version = ''
    $collector_install_dir = "${::win_programfiles}\\Splunk\\OpenTelemetry Collector"
    $collector_config_dir = "${::win_programdata}\\Splunk\\OpenTelemetry Collector"
    $splunk_bundle_dir = "${collector_install_dir}\\agent-bundle"
    $splunk_collectd_dir = "${splunk_bundle_dir}\\run\\collectd"
    $collector_config_source = "${collector_install_dir}\\agent_config.yaml"
    $collector_config_dest = "${collector_config_dir}\\agent_config.yaml"
    $fluentd_version = $fluentd_version_default
    $fluentd_config_source = "${collector_install_dir}\\fluentd\\td-agent.conf"
    $fluentd_config_dest = ''
  } else {
    fail("Your OS (${::osfamily}) is not supported by the Splunk OpenTelemetry Collector")
  }
}
