# Cookbook:: splunk-otel-collector
# Recipe:: collector_yum_repo

td_agent_major_version = node['splunk-otel-collector']['fluentd_version'].split('.')[0]

distro = if platform_family?('amazon')
           'amazon'
         else
           'redhat'
         end

yum_repository 'treasure-data' do
  description 'TreasureData Repository'
  baseurl "#{node['splunk-otel-collector']['fluentd_base_url']}/#{td_agent_major_version}/#{distro}/$releasever/$basearch"
  gpgcheck true
  gpgkey "#{node['splunk-otel-collector']['fluentd_base_url']}/GPG-KEY-td-agent"
  enabled true
  action :create
  notifies :run, 'execute[add-rhel-key]', :immediately
end

execute 'add-rhel-key' do
  command "rpm --import #{node['splunk-otel-collector']['fluentd_base_url']}/GPG-KEY-td-agent"
  action :nothing
end
