# Cookbook:: splunk_otel_collector
# Recipe:: default

ruby_block 'splunk-access-token-unset' do
  block do
    raise "Set ['splunk_otel_collector']['splunk_access_token'] as an attribute or on the node's run_state."
  end
  only_if { node['splunk_otel_collector']['splunk_access_token'].nil? && node.run_state['splunk_otel_collector']['splunk_access_token'].nil? }
end

if platform_family?('windows')
  include_recipe 'splunk_otel_collector::collector_win_config_options'
  include_recipe 'splunk_otel_collector::collector_win_install'

  # Older MSI versions can't properly setup the collector configuration
  # in this case, we need to use the registry to set the environment variables.
  # Custom variables also require using the registry, as the MSI doesn't support it.
  if !node['splunk_otel_collector']['collector_msi_is_configurable'] || !node['splunk_otel_collector']['collector_additional_env_vars'].empty?
    include_recipe 'splunk_otel_collector::collector_win_registry'
  end

  directory ::File.dirname(node['splunk_otel_collector']['collector_config_dest']) do
    action :create
  end

  template node['splunk_otel_collector']['collector_config_dest'] do
    source 'agent_config.yaml.erb'
    only_if { node['splunk_otel_collector']['collector_config'] != {} }
    notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  end

  remote_file node['splunk_otel_collector']['collector_config_dest'] do
    source node['splunk_otel_collector']['collector_config_source'].to_s
    only_if { ::File.exist?(node['splunk_otel_collector']['collector_config_source']) && node['splunk_otel_collector']['collector_config'] == {} }
    notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
  end

  windows_service 'splunk-otel-collector' do
    service_name node['splunk_otel_collector']['service_name']
    action [:enable, :start]
  end

elsif platform_family?('debian', 'rhel', 'amazon', 'suse')
  if platform_family?('debian')
    package %w(apt-transport-https gnupg)
    include_recipe 'splunk_otel_collector::collector_deb_repo'
  elsif platform_family?('rhel', 'amazon')
    package %w(libcap)
    include_recipe 'splunk_otel_collector::collector_yum_repo'
  elsif platform_family?('suse')
    package %w(libcap-progs)
    include_recipe 'splunk_otel_collector::collector_zypper_repo'
  end

  # splunk-otel-collector package should already be installed for local artifact testing
  unless node['splunk_otel_collector']['local_artifact_testing_enabled']
    package 'splunk-otel-collector' do
      action :install
      version node['splunk_otel_collector']['collector_version'] if node['splunk_otel_collector']['collector_version'] != 'latest'
      flush_cache [ :before ] if platform_family?('amazon', 'rhel')
      options '--allow-downgrades' if platform_family?('debian') \
        && node['packages'] \
        && node['packages']['apt'] \
        && Gem::Version.new(node['packages']['apt']['version'].split('~').first) >= Gem::Version.new('1.1.0')
      allow_downgrade true if platform_family?('amazon', 'rhel', 'suse')
      notifies :restart, 'service[splunk-otel-collector]', :delayed
    end
  end

  include_recipe 'splunk_otel_collector::collector_service_owner'

  directory ::File.dirname(node['splunk_otel_collector']['collector_config_dest']) do
    action :create
  end

  template node['splunk_otel_collector']['collector_config_dest'] do
    source 'agent_config.yaml.erb'
    owner node['splunk_otel_collector']['user']
    group node['splunk_otel_collector']['group']
    mode '0600'
    only_if { node['splunk_otel_collector']['collector_config'] != {} }
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  remote_file node['splunk_otel_collector']['collector_config_dest'] do
    source node['splunk_otel_collector']['collector_config_source'].to_s
    owner node['splunk_otel_collector']['user']
    group node['splunk_otel_collector']['group']
    mode '0600'
    only_if { node['splunk_otel_collector']['collector_config'] == {} }
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  template '/etc/otel/collector/splunk-otel-collector.conf' do
    source 'splunk-otel-collector.conf.erb'
    owner node['splunk_otel_collector']['user']
    group node['splunk_otel_collector']['group']
    mode '0600'
    notifies :restart, 'service[splunk-otel-collector]', :delayed
  end

  service 'splunk-otel-collector' do
    service_name node['splunk_otel_collector']['service_name']
    action [:enable, :start]
  end

  if node['splunk_otel_collector']['with_auto_instrumentation'].to_s.downcase == 'true'
    include_recipe 'splunk_otel_collector::auto_instrumentation'
  end
else
  raise "Platform family #{node['platform_family']} not supported."
end
