# Class for default param values based on OS
class splunk_otel_collector::params {
  $fluentd_version_default = '4.3.2'
  $collector_additional_env_vars = {}

  if $facts['os']['family'] == 'redhat' or $facts['os']['family'] == 'debian' or $facts['os']['family'] == 'suse' {
    $collector_version = 'latest'
    $collector_config_dir = '/etc/otel/collector'
    $fluentd_config_dir = "${collector_config_dir}/fluentd"
    $splunk_bundle_dir = '/usr/lib/splunk-otel-collector/agent-bundle'
    $splunk_collectd_dir = "${splunk_bundle_dir}/run/collectd"
    $collector_config_source = "${collector_config_dir}/agent_config.yaml"
    $collector_config_dest = $collector_config_source
    $fluentd_base_url = 'https://packages.treasuredata.com'
    if $facts['os']['family'] == 'debian' {
      $fluentd_version = "${fluentd_version_default}-1"
    } else {
      $fluentd_version = $fluentd_version_default
    }
    $fluentd_config_source = "${fluentd_config_dir}/fluent.conf"
    $fluentd_config_dest = $fluentd_config_source
    $auto_instrumentation_version = 'latest'
    $auto_instrumentation_java_agent_jar = '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar'
  } elsif $facts['os']['family'] == 'windows' {
    $collector_version = 'latest'
    $collector_install_dir = "${facts['win_programfiles']}\\Splunk\\OpenTelemetry Collector"
    $collector_config_dir = "${facts['win_programdata']}\\Splunk\\OpenTelemetry Collector"
    $splunk_bundle_dir = "${collector_install_dir}\\agent-bundle"
    $splunk_collectd_dir = "${splunk_bundle_dir}\\run\\collectd"
    $collector_config_source = "${collector_install_dir}\\agent_config.yaml"
    $default_win_config_file = $collector_config_source
    $collector_config_dest = "${collector_config_dir}\\agent_config.yaml"
    $fluentd_base_url = 'https://s3.amazonaws.com/packages.treasuredata.com'
    $fluentd_version = $fluentd_version_default
    $fluentd_config_source = "${collector_install_dir}\\fluentd\\td-agent.conf"
    $fluentd_config_dest = ''
    $auto_instrumentation_version = ''
    $auto_instrumentation_java_agent_jar = ''
  } else {
    fail("Your OS (${facts['os']['family']}) is not supported by the Splunk OpenTelemetry Collector")
  }
}
