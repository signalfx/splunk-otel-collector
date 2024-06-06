# Cookbook:: splunk_otel_collector
# Recipe:: collector_win_registry

collector_env_vars = node['splunk_otel_collector']['collector_win_env_vars']

collector_env_vars_strings = []
collector_env_vars.each do |item|
  collector_env_vars_strings |= [ "#{item[:name]}=#{item[:data]}" ]
end
collector_env_vars_strings.sort!
registry_key 'HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\splunk-otel-collector' do
  values [{
    name: 'Environment',
    type: :multi_string,
    data: collector_env_vars_strings,
  }]
  action :create
  notifies :restart, 'windows_service[splunk-otel-collector]', :delayed
end
