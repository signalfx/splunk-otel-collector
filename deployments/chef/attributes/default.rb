default['splunk-otel-collector']['repo_base_url'] = 'https://splunk.jfrog.io/splunk'
default['splunk-otel-collector']['package_stage'] = 'release'

default['splunk-otel-collector']['debian_repo_url'] = "#{node['splunk-otel-collector']['repo_base_url']}/otel-collector-deb"
default['splunk-otel-collector']['debian_gpg_key_url'] = "#{node['splunk-otel-collector']['debian_repo_url']}/splunk-B3CD4420.gpg"

default['splunk-otel-collector']['rhel_repo_url'] = "#{node['splunk-otel-collector']['repo_base_url']}/otel-collector-rpm"
default['splunk-otel-collector']['rhel_gpg_key_url'] = "#{node['splunk-otel-collector']['rhel_repo_url']}/splunk-B3CD4420.pub"

default['splunk-otel-collector']['windows_repo_url'] = "https://dl.signalfx.com/splunk-otel-collector/msi/#{node['splunk-otel-collector']['package_stage']}"

default['splunk-otel-collector']['service_name'] = 'splunk-otel-collector'

default['splunk-otel-collector']['splunk_access_token'] = nil
default['splunk-otel-collector']['splunk_realm'] = 'us0'

default['splunk-otel-collector']['splunk_api_url'] = "https://api.#{node['splunk-otel-collector']['splunk_realm']}.signalfx.com"
default['splunk-otel-collector']['splunk_ingest_url'] = "https://ingest.#{node['splunk-otel-collector']['splunk_realm']}.signalfx.com"
default['splunk-otel-collector']['splunk_trace_url'] = "#{node['splunk-otel-collector']['splunk_ingest_url']}/v2/trace"
default['splunk-otel-collector']['splunk_hec_url'] = "#{node['splunk-otel-collector']['splunk_ingest_url']}/v1/log"
default['splunk-otel-collector']['splunk_hec_token'] = "#{node['splunk-otel-collector']['splunk_access_token']}"
default['splunk-otel-collector']['splunk_memory_total_mib'] = '512'
default['splunk-otel-collector']['splunk_ballast_size_mib'] = ''

default['splunk-otel-collector']['collector_config'] = {}

default['splunk-otel-collector']['with_fluentd'] = true
default['splunk-otel-collector']['fluentd_base_url'] = 'https://packages.treasuredata.com'
default['splunk-otel-collector']['fluentd_version'] = if platform_family?('debian')
                                                        case node['lsb']['codename']
                                                        when 'stretch'
                                                          '3.7.1-0'
                                                        else
                                                          '4.3.0-1'
                                                        end
                                                      else
                                                        '4.3.0'
                                                      end

if platform_family?('windows')
  default['splunk-otel-collector']['collector_version'] = 'latest'
  default['splunk-otel-collector']['collector_version_url'] = "#{node['splunk-otel-collector']['windows_repo_url']}/#{node['splunk-otel-collector']['service_name']}/msi/#{node['splunk-otel-collector']['package_stage']}/latest.txt"

  collector_install_dir = "#{ENV['ProgramFiles']}\\Splunk\\OpenTelemetry Collector"

  default['splunk-otel-collector']['collector_config_source'] = 'file:///' + "#{collector_install_dir}\\agent_config.yaml"
  default['splunk-otel-collector']['collector_config_dest'] = "#{ENV['ProgramData']}\\Splunk\\OpenTelemetry Collector\\agent_config.yaml"
  default['splunk-otel-collector']['collector_version_file'] = "#{collector_install_dir}\\collector_version.txt"

  default['splunk-otel-collector']['splunk_bundle_dir'] = "#{collector_install_dir}\\agent-bundle"
  default['splunk-otel-collector']['splunk_collectd_dir'] = "#{node['splunk-otel-collector']['splunk_bundle_dir']}\\run\\collectd"

  default['splunk-otel-collector']['fluentd_config_source'] = 'file:///' + "#{collector_install_dir}\\fluentd\\td-agent.conf"
  default['splunk-otel-collector']['fluentd_config_dest'] = "#{ENV['SystemDrive']}\\opt\\td-agent\\etc\\td-agent\\td-agent.conf"
  default['splunk-otel-collector']['fluentd_version_file'] = "#{collector_install_dir}\\fluentd_version.txt"

elsif platform_family?('debian', 'rhel', 'amazon', 'suse')
  default['splunk-otel-collector']['collector_version'] = 'latest'

  default['splunk-otel-collector']['collector_config_source'] = 'file:///etc/otel/collector/agent_config.yaml'
  default['splunk-otel-collector']['collector_config_dest'] = '/etc/otel/collector/agent_config.yaml'

  default['splunk-otel-collector']['splunk_bundle_dir'] = '/usr/lib/splunk-otel-collector/agent-bundle'
  default['splunk-otel-collector']['splunk_collectd_dir'] = "#{node['splunk-otel-collector']['splunk_bundle_dir']}/run/collectd"

  default['splunk-otel-collector']['user'] = 'splunk-otel-collector'
  default['splunk-otel-collector']['group'] = 'splunk-otel-collector'

  default['splunk-otel-collector']['fluentd_config_source'] = 'file:///etc/otel/collector/fluentd/fluent.conf'
  default['splunk-otel-collector']['fluentd_config_dest'] = '/etc/otel/collector/fluentd/fluent.conf'

end
