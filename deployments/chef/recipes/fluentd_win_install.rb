# Cookbook:: splunk_otel_collector
# Recipe:: fluentd_win_install

td_agent_major_version = node['splunk_otel_collector']['fluentd_version'].split('.')[0]

remote_file "#{ENV['TEMP']}\\td-agent-#{node['splunk_otel_collector']['fluentd_version']}-x64.msi" do
  source "#{node['splunk_otel_collector']['fluentd_base_url']}/#{td_agent_major_version}/windows/td-agent-#{node['splunk_otel_collector']['fluentd_version']}-x64.msi"
  action :create
  only_if { !::File.exist?(node['splunk_otel_collector']['fluentd_version_file']) || (::File.readlines(node['splunk_otel_collector']['fluentd_version_file']).first.strip != node['splunk_otel_collector']['fluentd_version']) }
end

windows_package 'fluentd' do
  source "#{ENV['TEMP']}\\td-agent-#{node['splunk_otel_collector']['fluentd_version']}-x64.msi"
  action :install
  notifies :restart, 'windows_service[td-agent]', :delayed
  notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  only_if { !::File.exist?(node['splunk_otel_collector']['fluentd_version_file']) || (::File.readlines(node['splunk_otel_collector']['fluentd_version_file']).first.strip != node['splunk_otel_collector']['fluentd_version']) }
end

file node['splunk_otel_collector']['fluentd_version_file'] do
  content node['splunk_otel_collector']['fluentd_version']
end

directory ::File.dirname(node['splunk_otel_collector']['fluentd_config_dest']) do
  action :create
end

directory "#{ENV['SystemDrive']}\\opt\\td-agent\\etc\\td-agent\\conf.d" do
  action :create
end

remote_file node['splunk_otel_collector']['fluentd_config_dest'] do
  source "#{node['splunk_otel_collector']['fluentd_config_source']}"
  notifies :restart, 'windows_service[td-agent]', :delayed
end

remote_file "#{ENV['SystemDrive']}\\opt\\td-agent\\etc\\td-agent\\conf.d\\eventlog.conf" do
  source 'file:///' + "#{ENV['ProgramFiles']}\\Splunk\\OpenTelemetry Collector\\fluentd\\conf.d\\eventlog.conf"
  notifies :restart, 'windows_service[td-agent]', :delayed
end

windows_service 'td-agent' do
  service_name 'fluentdwinsvc'
  action [:enable, :start]
end
