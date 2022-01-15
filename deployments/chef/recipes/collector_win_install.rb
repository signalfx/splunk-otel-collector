#
# Cookbook:: splunk-otel-collector
# Recipe:: collector_win_install
#
# Copyright:: 2021, The Authors, All Rights Reserved.

remote_file "#{ENV['TEMP']}/latest.txt" do
  source "#{node['splunk-otel-collector']['windows_repo_url']}/latest.txt"
  action :create
  only_if { node['splunk-otel-collector']['collector_version'] == 'latest' }
end.run_action(:create)

collector_version = if node['splunk-otel-collector']['collector_version'] == 'latest'
                      ::File.read("#{ENV['TEMP']}/latest.txt")
                    else
                      node['splunk-otel-collector']['collector_version']
                    end

remote_file 'Download msi' do
  path "#{ENV['TEMP']}/splunk-otel-collector-#{collector_version}-amd64.msi"
  source "#{node['splunk-otel-collector']['windows_repo_url']}/splunk-otel-collector-#{collector_version}-amd64.msi"
  action :create
end

windows_package 'splunk-otel-collector' do
  source "#{ENV['TEMP']}/splunk-otel-collector-#{collector_version}-amd64.msi"
  action :install
  notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
end
