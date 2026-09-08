# Cookbook:: splunk_otel_collector
# Recipe:: collector_service_owner

require 'shellwords'

service_user = node['splunk_otel_collector']['user']
service_group = node['splunk_otel_collector']['group']
collector_directories = %w(/etc/otel/collector /var/lib/otelcol)

group service_group do
  system true
  not_if "getent group #{service_group}"
end

user service_user do
  system true
  manage_home false
  home '/etc/otel/collector'
  group service_group
  shell '/usr/sbin/nologin' if platform_family?('debian')
  shell '/sbin/nologin' if platform_family?('rhel', 'amazon', 'suse')
  not_if "getent passwd #{service_user}"
end

collector_directories.each do |path|
  directory "ensure #{path} exists" do
    path path
    mode '0755'
    action :create
  end
end

execute 'systemctl daemon-reload' do
  notifies :restart, 'service[splunk-otel-collector]', :delayed
  action :nothing
end

directory '/etc/systemd/system/splunk-otel-collector.service.d' do
  action :create
end

file '/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf' do
  content "[Service]\nUser=#{service_user}\nGroup=#{service_group}"
  mode '0644'
  notifies :stop, 'service[splunk-otel-collector]', :immediately
  notifies :run, 'execute[systemctl daemon-reload]', :immediately
  action :create
end

escaped_user = Shellwords.escape(service_user)
escaped_group = Shellwords.escape(service_group)
escaped_directories = collector_directories.map { |path| Shellwords.escape(path) }.join(' ')

execute 'set collector directory ownership recursively' do
  command "chown -R -- #{escaped_user}:#{escaped_group} #{escaped_directories}"
  only_if "find #{escaped_directories} \\( ! -user #{escaped_user} -o ! -group #{escaped_group} \\) " \
          '-print -quit | grep -q .'
end
