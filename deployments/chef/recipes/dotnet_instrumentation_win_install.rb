# Cookbook:: splunk_otel_collector
# Recipe:: dotnet_instrumentation_win_install

env_vars = [
  { name: 'COR_ENABLE_PROFILING', type: :string, data: 'true' },
  { name: 'COR_PROFILER', type: :string, data: '{B4C89B0F-9908-4F73-9F59-0D77C5A06874}' },
  { name: 'CORECLR_ENABLE_PROFILING', type: :string, data: 'true' },
  { name: 'CORECLR_PROFILER', type: :string, data: '{B4C89B0F-9908-4F73-9F59-0D77C5A06874}' },
  { name: 'SIGNALFX_ENV', type: :string, data: node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_environment'].to_s },
  { name: 'SIGNALFX_PROFILER_ENABLED', type: :string, data: node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_enable_profiler'].to_s.downcase },
  { name: 'SIGNALFX_PROFILER_MEMORY_ENABLED', type: :string, data: node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_enable_profiler_memory'].to_s.downcase },
  { name: 'SIGNALFX_SERVICE_NAME', type: :string, data: node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_service_name'].to_s },
]

node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_additional_options'].each do |key, value|
  env_vars.push({ name: key, type: :string, data: value.to_s })
end

iis_env_vars = []
env_vars.each do |item|
  iis_env_vars |= [ "#{item[:name]}=#{item[:data]}" ]
end

msi_name = node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_msi_url'].split('/')[-1]
msi_path = "#{ENV['TEMP']}\\#{msi_name}"

remote_file msi_path do
  source node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_msi_url']
  action :create
end

windows_package 'SignalFx .NET Tracing 64-bit' do
  source msi_path
  action :install
  notifies :run, 'powershell_script[iisreset]', :delayed
end

registry_key 'HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\W3SVC' do
  values [{
    name: 'Environment',
    type: :multi_string,
    data: iis_env_vars,
  }]
  action :create
  notifies :run, 'powershell_script[iisreset]', :delayed
end

registry_key 'HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment' do
  values env_vars
  action :create
  notifies :run, 'powershell_script[iisreset]', :delayed
  only_if { node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_system_wide'].to_s.downcase == 'true' }
end

powershell_script 'iisreset' do
  code <<-EOH
  try {
    Get-Command iisreset.exe
  } Catch {
    Exit
  }
  & { iisreset.exe }
  EOH
  action :nothing
  only_if { node['splunk_otel_collector']['signalfx_dotnet_auto_instrumentation_iisreset'].to_s.downcase == 'true' }
end
