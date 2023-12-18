describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

describe service('fluentdwinsvc') do
  it { should_not be_enabled }
  it { should_not be_running }
end

env_vars = [
  { name: 'COR_ENABLE_PROFILING', type: :string, data: 'true' },
  { name: 'COR_PROFILER', type: :string, data: '{B4C89B0F-9908-4F73-9F59-0D77C5A06874}' },
  { name: 'CORECLR_ENABLE_PROFILING', type: :string, data: 'true' },
  { name: 'CORECLR_PROFILER', type: :string, data: '{B4C89B0F-9908-4F73-9F59-0D77C5A06874}' },
  { name: 'SIGNALFX_ENV', type: :string, data: 'test-env' },
  { name: 'SIGNALFX_PROFILER_ENABLED', type: :string, data: 'true' },
  { name: 'SIGNALFX_PROFILER_MEMORY_ENABLED', type: :string, data: 'true' },
  { name: 'SIGNALFX_SERVICE_NAME', type: :string, data: 'test-service' },
  { name: 'MY_CUSTOM_OPTION1', type: :string, data: 'value1' },
  { name: 'MY_CUSTOM_OPTION2', type: :string, data: 'value2' },
]

describe registry_key('HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment') do
  its('SIGNALFX_DOTNET_TRACER_HOME') { should cmp "#{ENV['ProgramFiles']}\\SignalFx\\.NET Tracing\\" }
  env_vars.each do |item|
    its(item[:name]) { should eq item[:data] }
  end
end

iis_env = []
env_vars.each do |item|
  iis_env |= [ "#{item[:name]}=#{item[:data]}" ]
end

describe registry_key('HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\W3SVC') do
  it { should have_property_value('Environment', :multi_string, iis_env) }
end
