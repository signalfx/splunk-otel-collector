#
# Cookbook:: splunk-otel-collector
# Recipe:: collector_deb_repo
#
# Copyright:: 2021, The Authors, All Rights Reserved.

remote_file '/etc/apt/trusted.gpg.d/splunk.gpg' do
  source node['splunk-otel-collector']['debian_gpg_key_url']
  mode '0644'
  action :create
end

package %w(apt-transport-https gnupg)

file '/etc/apt/sources.list.d/splunk-otel-collector.list' do
  content "deb #{node['splunk-otel-collector']['debian_repo_url']} #{node['splunk-otel-collector']['package_stage']} main\n"
  mode '0644'
  notifies :update, 'apt_update[update apt cache]', :immediately
end

apt_update 'update apt cache' do
  action :nothing
end
