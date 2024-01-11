# Cookbook:: splunk_otel_collector
# Recipe:: collector_win_registry

collector_env_vars = [
  { name: 'SPLUNK_CONFIG', type: :string, data: node['splunk_otel_collector']['collector_config_dest'].to_s },
  { name: 'SPLUNK_ACCESS_TOKEN', type: :string, data: node['splunk_otel_collector']['splunk_access_token'].to_s },
  { name: 'SPLUNK_REALM', type: :string, data: node['splunk_otel_collector']['splunk_realm'].to_s },
  { name: 'SPLUNK_API_URL', type: :string, data: node['splunk_otel_collector']['splunk_api_url'].to_s },
  { name: 'SPLUNK_INGEST_URL', type: :string, data: node['splunk_otel_collector']['splunk_ingest_url'].to_s },
  { name: 'SPLUNK_TRACE_URL', type: :string, data: node['splunk_otel_collector']['splunk_trace_url'].to_s },
  { name: 'SPLUNK_HEC_URL', type: :string, data: node['splunk_otel_collector']['splunk_hec_url'].to_s },
  { name: 'SPLUNK_HEC_TOKEN', type: :string, data: node['splunk_otel_collector']['splunk_hec_token'].to_s },
  { name: 'SPLUNK_MEMORY_TOTAL_MIB', type: :string, data: node['splunk_otel_collector']['splunk_memory_total_mib'].to_s },
  { name: 'SPLUNK_BALLAST_SIZE_MIB', type: :string, data: node['splunk_otel_collector']['splunk_ballast_size_mib'].to_s },
  { name: 'SPLUNK_BUNDLE_DIR', type: :string, data: node['splunk_otel_collector']['splunk_bundle_dir'].to_s },
  { name: 'SPLUNK_COLLECTD_DIR', type: :string, data: node['splunk_otel_collector']['splunk_collectd_dir'].to_s },
]

unless node['splunk_otel_collector']['splunk_listen_interface'].to_s.strip.empty?
  collector_env_vars.push({ name: 'SPLUNK_LISTEN_INTERFACE', type: :string, data: node['splunk_otel_collector']['splunk_listen_interface'].to_s })
end

node['splunk_otel_collector']['collector_additional_env_vars'].each do |key, value|
  collector_env_vars.push({ name: key, type: :string, data: value.to_s })
end

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
