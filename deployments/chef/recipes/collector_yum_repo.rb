# Cookbook:: splunk-otel-collector
# Recipe:: collector_yum_repo

yum_repository 'splunk-otel-collector' do
  description 'Splunk OpenTelemetry Collector Repository'
  baseurl "#{node['splunk-otel-collector']['rhel_repo_url']}/#{node['splunk-otel-collector']['package_stage']}/$basearch/"
  gpgcheck true
  gpgkey node['splunk-otel-collector']['rhel_gpg_key_url']
  enabled true
  action :create
  notifies :run, 'execute[add-rhel-key]', :immediately
end

execute 'add-rhel-key' do
  command "rpm --import #{node['splunk-otel-collector']['rhel_gpg_key_url']}"
  action :nothing
end
