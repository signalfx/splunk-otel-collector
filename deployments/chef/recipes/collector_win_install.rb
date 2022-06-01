# Cookbook:: splunk_otel_collector
# Recipe:: collector_win_install

remote_file "#{ENV['TEMP']}/latest.txt" do
  source "#{node['splunk_otel_collector']['windows_repo_url']}/latest.txt"
  action :create
  only_if { node['splunk_otel_collector']['collector_version'] == 'latest' }
end.run_action(:create)

collector_version = if node['splunk_otel_collector']['collector_version'] == 'latest'
                      ::File.read("#{ENV['TEMP']}/latest.txt")
                    else
                      node['splunk_otel_collector']['collector_version']
                    end

remote_file 'Download msi' do
  path "#{ENV['TEMP']}/splunk-otel-collector-#{collector_version}-amd64.msi"
  source "#{node['splunk_otel_collector']['windows_repo_url']}/splunk-otel-collector-#{collector_version}-amd64.msi"
  action :create
  only_if { !::File.exist?(node['splunk_otel_collector']['collector_version_file']) || (::File.readlines(node['splunk_otel_collector']['collector_version_file']).first.strip != collector_version) }
end

windows_package 'splunk-otel-collector' do
  source "#{ENV['TEMP']}/splunk-otel-collector-#{collector_version}-amd64.msi"
  action :install
  notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  only_if { !::File.exist?(node['splunk_otel_collector']['collector_version_file']) || (::File.readlines(node['splunk_otel_collector']['collector_version_file']).first.strip != collector_version) }
end

file node['splunk_otel_collector']['collector_version_file'] do
  content collector_version
end
