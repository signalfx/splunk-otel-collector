libsplunk_path = '/usr/lib/splunk-instrumentation/libsplunk.so'
resource_attributes = 'splunk.zc.method=splunk-otel-auto-instrumentation-\d+\.\d+\.\d+'
dotnet_home = '/usr/lib/splunk-instrumentation/splunk-otel-dotnet'

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

describe file('/etc/splunk/zeroconfig/java.conf') do
  it { should_not exist }
end

describe file('/etc/splunk/zeroconfig/node.conf') do
  it { should_not exist }
end

describe file('/usr/lib/splunk-instrumentation/instrumentation.conf') do
  it { should_not exist }
end

describe file('/etc/splunk/zeroconfig/dotnet.conf') do
  its('content') { should match /^CORECLR_ENABLE_PROFILING=1$/ }
  its('content') { should match /^CORECLR_PROFILER=\{918728DD-259F-4A6A-AC2B-B85E1B658318\}$/ }
  its('content') { should match %r{^CORECLR_PROFILER_PATH=#{dotnet_home}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so$} }
  its('content') { should match %r{^DOTNET_ADDITIONAL_DEPS=#{dotnet_home}/AdditionalDeps$} }
  its('content') { should match %r{^DOTNET_SHARED_STORE=#{dotnet_home}/store$} }
  its('content') { should match %r{^DOTNET_STARTUP_HOOKS=#{dotnet_home}/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll$} }
  its('content') { should match /^OTEL_DOTNET_AUTO_HOME=#{dotnet_home}$/ }
  its('content') { should match /^OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation$/ }
  its('content') { should match /^OTEL_RESOURCE_ATTRIBUTES=#{resource_attributes}$/ }
  its('content') { should_not match /.*OTEL_SERVICE_NAME.*/ }
  its('content') { should match /^SPLUNK_PROFILER_ENABLED=false$/ }
  its('content') { should match /^SPLUNK_PROFILER_MEMORY_ENABLED=false$/ }
  its('content') { should match /^SPLUNK_METRICS_ENABLED=false$/ }
  its('content') { should_not match /.*OTEL_EXPORTER_OTLP_ENDPOINT.*/ }
  its('content') { should_not match /.*OTEL_EXPORTER_OTLP_PROTOCOL.*/ }
  its('content') { should_not match /.*OTEL_METRICS_EXPORTER.*/ }
  its('content') { should_not match /.*OTEL_LOGS_EXPORTER.*/ }
end

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

describe service('td-agent') do
  it { should_not be_enabled }
  it { should_not be_running }
end
