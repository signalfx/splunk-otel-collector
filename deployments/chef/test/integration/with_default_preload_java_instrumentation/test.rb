libsplunk_path = '/usr/lib/splunk-instrumentation/libsplunk.so'
java_tool_options = '-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar'
resource_attributes = 'splunk.zc.method=splunk-otel-auto-instrumentation-\d+\.\d+\.\d+'
otlp_endpoint = 'http://127.0.0.1:4317'

describe package('splunk-otel-auto-instrumentation') do
  it { should be_installed }
end

describe npm('@splunk/otel', path: '/usr/lib/splunk-instrumentation/splunk-otel-js') do
  it { should_not be_installed }
end

describe file('/etc/ld.so.preload') do
  its('content') { should match /^#{libsplunk_path}$/ }
end

describe file('/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf') do
  it { should_not exist }
end

describe file('/etc/splunk/zeroconfig/node.conf') do
  it { should_not exist }
end

describe file('/usr/lib/splunk-instrumentation/instrumentation.conf') do
  it { should_not exist }
end

describe file('/etc/splunk/zeroconfig/java.conf') do
  its('content') { should match /^JAVA_TOOL_OPTIONS=#{java_tool_options}$/ }
  its('content') { should match /^OTEL_RESOURCE_ATTRIBUTES=#{resource_attributes}$/ }
  its('content') { should_not match /.*OTEL_SERVICE_NAME.*/ }
  its('content') { should match /^SPLUNK_PROFILER_ENABLED=false$/ }
  its('content') { should match /^SPLUNK_PROFILER_MEMORY_ENABLED=false$/ }
  its('content') { should match /^SPLUNK_METRICS_ENABLED=false$/ }
  its('content') { should match /^OTEL_EXPORTER_OTLP_ENDPOINT=#{otlp_endpoint}$/ }
end

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

describe service('td-agent') do
  it { should_not be_enabled }
  it { should_not be_running }
end
