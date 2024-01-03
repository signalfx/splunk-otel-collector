default['splunk_otel_collector']['repo_base_url'] = 'https://splunk.jfrog.io/splunk'
default['splunk_otel_collector']['package_stage'] = 'release'

default['splunk_otel_collector']['debian_repo_url'] = "#{node['splunk_otel_collector']['repo_base_url']}/otel-collector-deb"
default['splunk_otel_collector']['debian_gpg_key_url'] = "#{node['splunk_otel_collector']['debian_repo_url']}/splunk-B3CD4420.gpg"

default['splunk_otel_collector']['rhel_repo_url'] = "#{node['splunk_otel_collector']['repo_base_url']}/otel-collector-rpm"
default['splunk_otel_collector']['rhel_gpg_key_url'] = "#{node['splunk_otel_collector']['rhel_repo_url']}/splunk-B3CD4420.pub"

default['splunk_otel_collector']['windows_repo_url'] = "https://dl.signalfx.com/splunk-otel-collector/msi/#{node['splunk_otel_collector']['package_stage']}"

default['splunk_otel_collector']['service_name'] = 'splunk-otel-collector'

default['splunk_otel_collector']['splunk_access_token'] = nil
default['splunk_otel_collector']['splunk_realm'] = 'us0'

default['splunk_otel_collector']['splunk_api_url'] = "https://api.#{node['splunk_otel_collector']['splunk_realm']}.signalfx.com"
default['splunk_otel_collector']['splunk_ingest_url'] = "https://ingest.#{node['splunk_otel_collector']['splunk_realm']}.signalfx.com"
default['splunk_otel_collector']['splunk_trace_url'] = "#{node['splunk_otel_collector']['splunk_ingest_url']}/v2/trace"
default['splunk_otel_collector']['splunk_hec_url'] = "#{node['splunk_otel_collector']['splunk_ingest_url']}/v1/log"
default['splunk_otel_collector']['splunk_hec_token'] = "#{node['splunk_otel_collector']['splunk_access_token']}"
default['splunk_otel_collector']['splunk_memory_total_mib'] = '512'
default['splunk_otel_collector']['splunk_ballast_size_mib'] = ''
default['splunk_otel_collector']['splunk_listen_interface'] = ''

default['splunk_otel_collector']['collector_config'] = {}

default['splunk_otel_collector']['with_fluentd'] = false
default['splunk_otel_collector']['fluentd_version'] = if platform_family?('debian')
                                                        case node['lsb']['codename']
                                                        when 'stretch'
                                                          '3.7.1-0'
                                                        else
                                                          '4.3.2-1'
                                                        end
                                                      else
                                                        '4.3.2'
                                                      end
default['splunk_otel_collector']['collector_additional_env_vars'] = {}

if platform_family?('windows')
  default['splunk_otel_collector']['collector_version'] = 'latest'
  default['splunk_otel_collector']['collector_version_url'] = "#{node['splunk_otel_collector']['windows_repo_url']}/#{node['splunk_otel_collector']['service_name']}/msi/#{node['splunk_otel_collector']['package_stage']}/latest.txt"

  collector_install_dir = "#{ENV['ProgramFiles']}\\Splunk\\OpenTelemetry Collector"

  default['splunk_otel_collector']['collector_config_source'] = 'file:///' + "#{collector_install_dir}\\agent_config.yaml"
  default['splunk_otel_collector']['collector_config_dest'] = "#{ENV['ProgramData']}\\Splunk\\OpenTelemetry Collector\\agent_config.yaml"
  default['splunk_otel_collector']['collector_version_file'] = "#{collector_install_dir}\\collector_version.txt"

  default['splunk_otel_collector']['splunk_bundle_dir'] = "#{collector_install_dir}\\agent-bundle"
  default['splunk_otel_collector']['splunk_collectd_dir'] = "#{node['splunk_otel_collector']['splunk_bundle_dir']}\\run\\collectd"

  default['splunk_otel_collector']['fluentd_base_url'] = 'https://s3.amazonaws.com/packages.treasuredata.com'
  default['splunk_otel_collector']['fluentd_config_source'] = 'file:///' + "#{collector_install_dir}\\fluentd\\td-agent.conf"
  default['splunk_otel_collector']['fluentd_config_dest'] = "#{ENV['SystemDrive']}\\opt\\td-agent\\etc\\td-agent\\td-agent.conf"
  default['splunk_otel_collector']['fluentd_version_file'] = "#{collector_install_dir}\\fluentd_version.txt"

  default['splunk_otel_collector']['with_signalfx_dotnet_auto_instrumentation'] = false
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_version'] = '1.1.0'

  signalfx_dotnet_base_url = 'https://github.com/signalfx/signalfx-dotnet-tracing/releases/download'
  signalfx_dotnet_version = node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_version'].sub(/^v/, '')
  signalfx_dotnet_msi_name = "signalfx-dotnet-tracing-#{signalfx_dotnet_version}-x64.msi"
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_msi_url'] = "#{signalfx_dotnet_base_url}/v#{signalfx_dotnet_version}/#{signalfx_dotnet_msi_name}"

  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_iisreset'] = true
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_system_wide'] = false
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_environment'] = ''
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_service_name'] = ''
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_enable_profiler'] = false
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_enable_profiler_memory'] = false
  default['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_additional_options'] = {}
elsif platform_family?('debian', 'rhel', 'amazon', 'suse')
  default['splunk_otel_collector']['collector_version'] = 'latest'

  default['splunk_otel_collector']['collector_config_source'] = 'file:///etc/otel/collector/agent_config.yaml'
  default['splunk_otel_collector']['collector_config_dest'] = '/etc/otel/collector/agent_config.yaml'

  default['splunk_otel_collector']['splunk_bundle_dir'] = '/usr/lib/splunk-otel-collector/agent-bundle'
  default['splunk_otel_collector']['splunk_collectd_dir'] = "#{node['splunk_otel_collector']['splunk_bundle_dir']}/run/collectd"

  default['splunk_otel_collector']['user'] = 'splunk-otel-collector'
  default['splunk_otel_collector']['group'] = 'splunk-otel-collector'

  default['splunk_otel_collector']['fluentd_base_url'] = 'https://packages.treasuredata.com'
  default['splunk_otel_collector']['fluentd_config_source'] = 'file:///etc/otel/collector/fluentd/fluent.conf'
  default['splunk_otel_collector']['fluentd_config_dest'] = '/etc/otel/collector/fluentd/fluent.conf'

  default['splunk_otel_collector']['with_auto_instrumentation'] = false
  default['splunk_otel_collector']['auto_instrumentation_version'] = 'latest'
  default['splunk_otel_collector']['auto_instrumentation_systemd'] = false
  default['splunk_otel_collector']['auto_instrumentation_ld_so_preload'] = ''
  default['splunk_otel_collector']['auto_instrumentation_java_agent_jar'] = '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar'
  default['splunk_otel_collector']['auto_instrumentation_resource_attributes'] = ''
  default['splunk_otel_collector']['auto_instrumentation_service_name'] = ''
  default['splunk_otel_collector']['auto_instrumentation_generate_service_name'] = true
  default['splunk_otel_collector']['auto_instrumentation_disable_telemetry'] = false
  default['splunk_otel_collector']['auto_instrumentation_enable_profiler'] = false
  default['splunk_otel_collector']['auto_instrumentation_enable_profiler_memory'] = false
  default['splunk_otel_collector']['auto_instrumentation_enable_metrics'] = false
  default['splunk_otel_collector']['auto_instrumentation_otlp_endpoint'] = 'http://127.0.0.1:4317'
  default['splunk_otel_collector']['with_auto_instrumentation_sdks'] = %w(java nodejs)
  default['splunk_otel_collector']['auto_instrumentation_npm_path'] = 'npm'
end
