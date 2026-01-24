# Class for default param values based on OS
class splunk_otel_collector::params {
  $collector_additional_env_vars = {}

  if $facts['os']['family'] == 'redhat' or $facts['os']['family'] == 'debian' or $facts['os']['family'] == 'suse' {
    $collector_version = 'latest'
    $collector_config_dir = '/etc/otel/collector'
    $splunk_bundle_dir = '/usr/lib/splunk-otel-collector/agent-bundle'
    $splunk_collectd_dir = "${splunk_bundle_dir}/run/collectd"
    $collector_config_source = "${collector_config_dir}/agent_config.yaml"
    $collector_config_dest = $collector_config_source
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
    $auto_instrumentation_version = ''
    $auto_instrumentation_java_agent_jar = ''
  } else {
    fail("Your OS (${facts['os']['family']}) is not supported by the Splunk OpenTelemetry Collector")
  }
}
