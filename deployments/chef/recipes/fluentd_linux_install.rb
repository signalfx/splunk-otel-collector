# Cookbook:: splunk_otel_collector
# Recipe:: fluentd_linux_install

package 'td-agent' do
  action :install
  version node['splunk_otel_collector']['fluentd_version']
  flush_cache [ :before ] if platform_family?('amazon', 'rhel')
  options '--allow-downgrades' if platform_family?('debian') \
    && node['packages'] \
    && node['packages']['apt'] \
    && Gem::Version.new(node['packages']['apt']['version'].split('~')[0]) >= Gem::Version.new('1.1.0')

  allow_downgrade true if platform_family?('rhel', 'amazon')
  notifies :restart, 'service[td-agent]', :delayed
  notifies :restart, 'service[splunk-otel-collector]', :delayed
  notifies :run, 'execute[devtools]', :immediately if platform_family?('amazon', 'rhel')
end

directory '/etc/systemd/system/td-agent.service.d' do
  action :create
end

file '/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf' do
  content "[Service]\nEnvironment=FLUENT_CONF=#{node['splunk_otel_collector']['fluentd_config_dest']}"
  mode '0644'
  notifies :run, 'execute[systemctl daemon-reload td-agent]', :immediately
  action :create
end

# Install dependencies for plugins
if platform_family?('debian')
  package %w(build-essential libcap-ng0 libcap-ng-dev pkg-config)
else
  execute 'devtools' do
    command 'yum -y groupinstall "Development Tools"'
    action :nothing
  end
  package %w(libcap-ng libcap-ng-devel pkgconfig)
end

gem_package 'capng_c' do
  gem_binary 'td-agent-gem'
  version '0.2.2'
end

gem_package 'fluent-plugin-systemd' do
  gem_binary 'td-agent-gem'
  version '1.0.1'
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

directory ::File.dirname(node['splunk_otel_collector']['fluentd_config_dest']) do
  action :create
end

remote_file node['splunk_otel_collector']['fluentd_config_dest'] do
  source "#{node['splunk_otel_collector']['fluentd_config_source']}"
  owner 'td-agent'
  group 'td-agent'
  mode '0644'
  notifies :restart, 'service[td-agent]', :delayed
end
