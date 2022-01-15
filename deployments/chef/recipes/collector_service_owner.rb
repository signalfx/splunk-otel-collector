#
# Cookbook:: splunk-otel-collector
# Recipe:: collector_service_owner
#
# Copyright:: 2021, The Authors, All Rights Reserved.

group node['splunk-otel-collector']['group'] do
  system true
  not_if "getent group #{node['splunk-otel-collector']['group']}"
end

user node['splunk-otel-collector']['user'] do
  system true
  manage_home false
  home '/etc/otel/collector'
  group node['splunk-otel-collector']['group']
  shell '/usr/sbin/nologin' if platform_family?('debian')
  shell '/sbin/nologin' if platform_family?('rhel', 'amazon', 'suse')
  not_if "getent passwd #{node['splunk-otel-collector']['user']}"
end

execute 'init-tmpfile' do
  command 'systemd-tmpfiles --create --remove /etc/tmpfiles.d/splunk-otel-collector.conf'
  notifies :restart, 'service[splunk-otel-collector]', :delayed
  action :nothing
end

execute 'systemctl daemon-reload' do
  notifies :restart, 'service[splunk-otel-collector]', :delayed
  action :nothing
end

directory '/etc/systemd/system/splunk-otel-collector.service.d' do
  action :create
end

file '/etc/tmpfiles.d/splunk-otel-collector.conf' do
  content "D /run/splunk-otel-collector 0755 #{node['splunk-otel-collector']['user']} #{node['splunk-otel-collector']['group']} - -"
  mode '0644'
  notifies :run, 'execute[init-tmpfile]', :immediately
  action :create
end

file '/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf' do
  content "[Service]\nUser=#{node['splunk-otel-collector']['user']}\nGroup=#{node['splunk-otel-collector']['group']}"
  mode '0644'
  notifies :run, 'execute[systemctl daemon-reload]', :immediately
  action :create
end
