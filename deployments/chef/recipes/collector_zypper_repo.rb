# Cookbook:: splunk_otel_collector
# Recipe:: collector_zypper_repo

zypper_repository 'splunk-otel-collector' do
  description 'Splunk OpenTelemetry Collector Repository'
  baseurl "#{node['splunk_otel_collector']['rhel_repo_url']}/#{node['splunk_otel_collector']['package_stage']}/$basearch/"
  gpgcheck true
  gpgkey node['splunk_otel_collector']['rhel_gpg_key_url']
  type 'yum'
  enabled true
  action :create
  notifies :run, 'execute[add-rhel-key]', :immediately
end

execute 'add-rhel-key' do
  command "rpm --import #{node['splunk_otel_collector']['rhel_gpg_key_url']}"
  action :nothing
end
