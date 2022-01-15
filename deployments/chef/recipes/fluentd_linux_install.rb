#
# Cookbook:: splunk-otel-collector
# Recipe:: fluentd_linux_install
#
# Copyright:: 2021, The Authors, All Rights Reserved.

package 'td-agent' do
  action :install
  version node['splunk-otel-collector']['fluentd_version']
  flush_cache [ :before ] if platform_family?('rhel')
  options '--allow-downgrades' if platform_family?('debian') \
    && node['packages'] \
    && node['packages']['apt'] \
    && Gem::Version.new(node['packages']['apt']['version'].split('~')[0]) >= Gem::Version.new('1.1.0')

  allow_downgrade true if platform_family?('rhel', 'amazon')
  notifies :restart, 'service[td-agent]', :delayed
  notifies :restart, 'service[splunk-otel-collector]', :delayed
end

directory '/etc/systemd/system/td-agent.service.d' do
  action :create
end

file '/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf' do
  content "[Service]\nEnvironment=FLUENT_CONF=#{node['splunk-otel-collector']['fluentd_config_dest']}"
  mode '0644'
  notifies :run, 'execute[systemctl daemon-reload td-agent]', :immediately
  action :create
end

execute 'Install capng_c fluentd plugin' do
  command 'td-agent-gem install capng_c -v 0.2.2'
end

execute 'Install FluentD systemd plugin' do
  command 'td-agent-gem install fluent-plugin-systemd -v 1.0.1'
end

execute 'systemctl daemon-reload td-agent' do
  command 'systemctl daemon-reload'
  notifies :restart, 'service[td-agent]', :delayed
  action :nothing
end

service 'td-agent' do
  service_name 'td-agent'
  action [:enable, :start]
end

directory ::File.dirname(node['splunk-otel-collector']['fluentd_config_dest']) do
  action :create
end

remote_file node['splunk-otel-collector']['fluentd_config_dest'] do
  source "#{node['splunk-otel-collector']['fluentd_config_source']}"
  owner 'td-agent'
  group 'td-agent'
  mode '0644'
  notifies :restart, 'service[td-agent]', :delayed
end
