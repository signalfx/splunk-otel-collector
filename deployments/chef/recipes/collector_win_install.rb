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

remote_destination_path = "#{ENV['TEMP']}/splunk-otel-collector-#{collector_version}-amd64.msi"

if node['splunk_otel_collector']['local_artifact_testing_enabled']
  cookbook_file remote_destination_path do
    source 'splunk-otel-collector.msi'
    mode '0644'
  end
else
  remote_file 'Download msi' do
    path remote_destination_path
    source "#{node['splunk_otel_collector']['windows_repo_url']}/splunk-otel-collector-#{collector_version}-amd64.msi"
    action :create
    only_if { !::File.exist?(node['splunk_otel_collector']['collector_version_file']) || (::File.readlines(node['splunk_otel_collector']['collector_version_file']).first.strip != collector_version) }
  end
end

msi_is_configurable = Gem::Version.new(collector_version) >= Gem::Version.new('0.98.0')
node.default['splunk_otel_collector']['collector_msi_is_configurable'] = msi_is_configurable
msi_install_properties = node['splunk_otel_collector']['collector_win_env_vars']
                         .reject { |item| item[:data].nil? || item[:data] == '' }
                         .map { |item| "#{item[:name]}=\"#{item[:data]}\"" }
                         .join(' ')

collector_cmd_line_configurable = Gem::Version.new(collector_version) >= Gem::Version.new('0.127.0')
if collector_cmd_line_configurable && !node['splunk_otel_collector']['collector_command_line_args'].strip.empty?
  msi_install_properties += " COLLECTOR_SVC_ARGS=\"#{node['splunk_otel_collector']['collector_command_line_args']}\""
end

windows_package 'splunk-otel-collector' do
  source remote_destination_path
  options msi_install_properties # If the MSI is not configurable, this will be ignored during installation.
  action :install
  notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  only_if { !::File.exist?(node['splunk_otel_collector']['collector_version_file']) || (::File.readlines(node['splunk_otel_collector']['collector_version_file']).first.strip != collector_version) }
end

file node['splunk_otel_collector']['collector_version_file'] do
  content collector_version
end
