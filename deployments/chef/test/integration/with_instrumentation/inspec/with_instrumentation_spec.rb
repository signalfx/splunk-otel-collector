libsplunk_path = '/usr/lib/splunk-instrumentation/libsplunk.so'
java_agent_path = '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar'
resource_attribute = 'deployment.environment=test'
service_name = 'test'

describe package('splunk-otel-auto-instrumentation') do
  it { should be_installed }
end

describe file('/etc/ld.so.preload') do
  its('content') { should match /^#{libsplunk_path}$/ }
end

describe file('/usr/lib/splunk-instrumentation/instrumentation.conf') do
  its('content') { should match /^java_agent_jar=#{java_agent_path}$/ }
  its('content') { should match /^resource_attributes=#{resource_attribute}$/ }
  its('content') { should match /^service_name=#{service_name}$/ }
  its('content') { should match /^generate_service_name=false$/ }
  its('content') { should match /^disable_telemetry=true$/ }
  its('content') { should match /^enable_profiler=true$/ }
  its('content') { should match /^enable_profiler_memory=true$/ }
  its('content') { should match /^enable_metrics=true$/ }
end

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

describe service('td-agent') do
  it { should_not be_enabled }
  it { should_not be_running }
end
