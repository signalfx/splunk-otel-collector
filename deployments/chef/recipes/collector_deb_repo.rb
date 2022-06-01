# Cookbook:: splunk_otel_collector
# Recipe:: collector_deb_repo

remote_file '/etc/apt/trusted.gpg.d/splunk.gpg' do
  source node['splunk_otel_collector']['debian_gpg_key_url']
  mode '0644'
  action :create
end

file '/etc/apt/sources.list.d/splunk-otel-collector.list' do
  content "deb #{node['splunk_otel_collector']['debian_repo_url']} #{node['splunk_otel_collector']['package_stage']} main\n"
  mode '0644'
  notifies :update, 'apt_update[update apt cache]', :immediately
end

apt_update 'update apt cache' do
  action :nothing
end
