# Cookbook:: splunk_otel_collector
# Recipe:: fluentd_deb_repo

td_agent_major_version = node['splunk_otel_collector']['fluentd_version'].split('.')[0]
codename = node['lsb']['codename']
distro = node['platform']

apt_repository 'treasure-data' do
  uri "#{node['splunk_otel_collector']['fluentd_base_url']}/#{td_agent_major_version}/#{distro}/#{codename}"
  arch 'amd64'
  distribution codename
  components ['contrib']
  key "#{node['splunk_otel_collector']['fluentd_base_url']}/GPG-KEY-td-agent"
  action :add
end
