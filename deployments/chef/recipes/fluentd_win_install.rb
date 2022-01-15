#
# Cookbook:: splunk-otel-collector
# Recipe:: fluentd_win_install
#
# Copyright:: 2021, The Authors, All Rights Reserved.

td_agent_major_version = node['splunk-otel-collector']['fluentd_version'].split('.')[0]

remote_file "#{ENV['TEMP']}\\td-agent-#{node['splunk-otel-collector']['fluentd_version']}-x64.msi" do
  source "#{node['splunk-otel-collector']['fluentd_base_url']}/#{td_agent_major_version}/windows/td-agent-#{node['splunk-otel-collector']['fluentd_version']}-x64.msi"
  action :create
end

windows_package 'fluentd' do
  source "#{ENV['TEMP']}\\td-agent-#{node['splunk-otel-collector']['fluentd_version']}-x64.msi"
  action :install
  notifies :restart, 'windows_service[td-agent]', :delayed
  notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
end

directory ::File.dirname(node['splunk-otel-collector']['fluentd_config_dest']) do
  action :create
end

directory "#{ENV['SystemDrive']}\\opt\\td-agent\\etc\\td-agent\\conf.d" do
  action :create
end

remote_file node['splunk-otel-collector']['fluentd_config_dest'] do
  source "#{node['splunk-otel-collector']['fluentd_config_source']}"
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
