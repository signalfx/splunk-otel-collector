#
# Cookbook:: splunk-otel-collector
# Recipe:: default
#
# Copyright:: 2021, The Authors, All Rights Reserved.

ruby_block 'splunk-access-token-unset' do
  block do
    raise "Set ['splunk_access_token']['splunk_access_token'] as an attribute or on the node's run_state."
  end
  only_if { node['splunk-otel-collector']['splunk_access_token'].nil? }
end

ruby_block 'splunk-realm-unset' do
  block do
    raise "Set ['splunk_realm']['splunk_realm'] as an attribute or on the node's run_state."
  end
  only_if { node['splunk-otel-collector']['splunk_realm'].nil? }
end

if platform_family?('windows')
  include_recipe 'splunk-otel-collector::collector_win_install'
  include_recipe 'splunk-otel-collector::collector_win_registry'

  directory ::File.dirname(node['splunk-otel-collector']['collector_config_dest']) do
    action :create
  end

  template node['splunk-otel-collector']['collector_config_dest'] do
    source 'agent_config.yaml.erb'
    only_if { node['splunk-otel-collector']['collector_config'] != {} }
    notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  end

  remote_file node['splunk-otel-collector']['collector_config_dest'] do
    source "#{node['splunk-otel-collector']['collector_config_source']}"
    only_if { node['splunk-otel-collector']['collector_config'] == {} }
    notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  end

  windows_service 'splunk-otel-collector' do
    service_name node['splunk-otel-collector']['service_name']
    action [:enable, :start]
  end

  if node['splunk-otel-collector']['with_fluentd'] != false
    include_recipe 'splunk-otel-collector::fluentd_win_install'
  end
elsif platform_family?('debian', 'rhel', 'amazon', 'suse')
  if platform_family?('debian')
    include_recipe 'splunk-otel-collector::collector_deb_repo'
  elsif platform_family?('rhel', 'amazon', 'suse')
    include_recipe 'splunk-otel-collector::collector_yum_repo'
  end

  package 'splunk-otel-collector' do
    action :install
    version node['splunk-otel-collector']['collector_version'] if node['splunk-otel-collector']['collector_version'] != 'latest'
    flush_cache [ :before ] if platform_family?('rhel')
    options '--allow-downgrades' if platform_family?('debian') \
      && node['packages'] \
      && node['packages']['apt'] \
      && Gem::Version.new(node['packages']['apt']['version'].split('~')[0]) >= Gem::Version.new('1.1.0')

    allow_downgrade true if platform_family?('rhel', 'amazon')
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  include_recipe 'splunk-otel-collector::collector_service_owner'

  directory ::File.dirname(node['splunk-otel-collector']['collector_config_dest']) do
    action :create
  end

  template node['splunk-otel-collector']['collector_config_dest'] do
    source 'agent_config.yaml.erb'
    owner node['splunk-otel-collector']['user']
    group node['splunk-otel-collector']['group']
    mode '0600'
    only_if { node['splunk-otel-collector']['collector_config'] != {} }
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  remote_file node['splunk-otel-collector']['collector_config_dest'] do
    source "#{node['splunk-otel-collector']['collector_config_source']}"
    owner node['splunk-otel-collector']['user']
    group node['splunk-otel-collector']['group']
    mode '0600'
    only_if { node['splunk-otel-collector']['collector_config'] == {} }
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  template '/etc/otel/collector/splunk-otel-collector.conf' do
    source 'splunk-otel-collector.conf.erb'
    owner node['splunk-otel-collector']['user']
    group node['splunk-otel-collector']['group']
    mode '0600'
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  service 'splunk-otel-collector' do
    service_name node['splunk-otel-collector']['service_name']
    action [:enable, :start]
  end

  if node['splunk-otel-collector']['with_fluentd'] != false
    if platform_family?('debian')
      include_recipe 'splunk-otel-collector::fluentd_deb_repo'
    elsif platform_family?('rhel', 'amazon')
      include_recipe 'splunk-otel-collector::fluentd_yum_repo'
    end
    if platform_family?('debian', 'rhel', 'amazon')
      include_recipe 'splunk-otel-collector::fluentd_linux_install'
    end
  end
end
