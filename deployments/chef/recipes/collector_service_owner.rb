# Cookbook:: splunk_otel_collector
# Recipe:: collector_service_owner

group node['splunk_otel_collector']['group'] do
  system true
  not_if "getent group #{node['splunk_otel_collector']['group']}"
end

user node['splunk_otel_collector']['user'] do
  system true
  manage_home false
  home '/etc/otel/collector'
  group node['splunk_otel_collector']['group']
  shell '/usr/sbin/nologin' if platform_family?('debian')
  shell '/sbin/nologin' if platform_family?('rhel', 'amazon', 'suse')
  not_if "getent passwd #{node['splunk_otel_collector']['user']}"
end

execute 'systemctl daemon-reload' do
  notifies :restart, 'service[splunk-otel-collector]', :delayed
  action :nothing
end

directory '/etc/systemd/system/splunk-otel-collector.service.d' do
  action :create
end

file '/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf' do
  content "[Service]\nUser=#{node['splunk_otel_collector']['user']}\nGroup=#{node['splunk_otel_collector']['group']}"
  mode '0644'
  notifies :run, 'execute[systemctl daemon-reload]', :immediately
  action :create
end
