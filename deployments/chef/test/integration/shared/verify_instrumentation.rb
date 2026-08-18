LIBOTELINJECT_PATH = '/usr/lib/splunk-instrumentation/libotelinject.so'.freeze
INJECTOR_CONFIG_PATH = '/etc/opentelemetry/injector/injector.conf'.freeze
INJECTOR_DEFAULT_ENV_PATH = '/etc/opentelemetry/injector/default_env.conf'.freeze
SYSTEMD_INSTRUMENTATION_PATH = '/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf'.freeze
INSTRUMENTATION_VERSION_PATTERN = '\d+\.\d+\.\d+(?:[-+._~][A-Za-z0-9.+_~-]*)?'.freeze
DOTNET_PLUGIN = 'Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation'.freeze

def verify_instrumentation(sdks:, systemd:, custom: false, custom_preload: false)
  describe package('splunk-otel-auto-instrumentation') do
    it { should be_installed }
  end

  if sdks.include?('nodejs')
    describe npm('@splunk/otel', path: '/usr/lib/splunk-instrumentation/splunk-otel-js') do
      it { should be_installed }
    end
  end

  describe file('/etc/ld.so.preload') do
    if systemd
      its('content') { should_not match(/^#{LIBOTELINJECT_PATH}$/) }
    else
      its('content') { should match(/^#{LIBOTELINJECT_PATH}$/) }
    end
    if custom_preload
      its('content') { should match(/^# my extra library$/) }
    else
      its('content') { should_not match(/^# my extra library$/) }
    end
  end

  [
    '/usr/lib/splunk-instrumentation/instrumentation.conf',
    '/etc/splunk/zeroconfig/java.conf',
    '/etc/splunk/zeroconfig/node.conf',
    '/etc/splunk/zeroconfig/dotnet.conf',
  ].each do |legacy_config|
    describe file(legacy_config) do
      it { should_not exist }
    end
  end

  describe file(INJECTOR_CONFIG_PATH) do
    its('content') do
      should match %r{^jvm_auto_instrumentation_agent_path=/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar$}
    end
    its('content') do
      should match %r{^nodejs_auto_instrumentation_agent_path=/usr/lib/splunk-instrumentation/splunk-otel-js/node_modules/@splunk/otel/instrument.js$}
    end
    its('content') do
      should match %r{^dotnet_auto_instrumentation_agent_path_prefix=/usr/lib/splunk-instrumentation/splunk-otel-dotnet$}
    end
    { 'java' => 'jvm', 'nodejs' => 'nodejs', 'dotnet' => 'dotnet' }.each do |sdk, injector_name|
      if sdks.include?(sdk)
        its('content') { should_not match(/^auto_instrumentation_disabled=#{injector_name}$/) }
      else
        its('content') { should match(/^auto_instrumentation_disabled=#{injector_name}$/) }
      end
    end
  end

  method = "splunk.zc.method=splunk-otel-auto-instrumentation-#{INSTRUMENTATION_VERSION_PATTERN}"
  method += '-systemd' if systemd
  method += ',deployment.environment.name=test' if custom
  describe file(INJECTOR_DEFAULT_ENV_PATH) do
    if sdks.include?('dotnet')
      its('content') { should match(/^OTEL_DOTNET_AUTO_PLUGINS=#{DOTNET_PLUGIN}$/) }
    else
      its('content') { should_not match(/^OTEL_DOTNET_AUTO_PLUGINS=/) }
    end
    its('content') { should match(/^OTEL_RESOURCE_ATTRIBUTES=#{method}$/) }
    its('content') { should match(/^SPLUNK_PROFILER_ENABLED=#{custom}$/) }
    its('content') { should match(/^SPLUNK_PROFILER_MEMORY_ENABLED=#{custom}$/) }
    its('content') { should match(/^SPLUNK_METRICS_ENABLED=#{custom}$/) }
    if custom
      its('content') { should match(/^OTEL_SERVICE_NAME=test$/) }
      its('content') { should match %r{^OTEL_EXPORTER_OTLP_ENDPOINT=http://0.0.0.0:4317$} }
      its('content') { should match(/^OTEL_EXPORTER_OTLP_PROTOCOL=grpc$/) }
      its('content') { should match(/^OTEL_METRICS_EXPORTER=none$/) }
      its('content') { should match(/^OTEL_LOGS_EXPORTER=none$/) }
    else
      its('content') { should_not match(/^OTEL_SERVICE_NAME=/) }
      its('content') { should_not match(/^OTEL_EXPORTER_OTLP_ENDPOINT=/) }
      its('content') { should_not match(/^OTEL_EXPORTER_OTLP_PROTOCOL=/) }
      its('content') { should_not match(/^OTEL_METRICS_EXPORTER=/) }
      its('content') { should_not match(/^OTEL_LOGS_EXPORTER=/) }
    end
  end

  describe file(SYSTEMD_INSTRUMENTATION_PATH) do
    if systemd
      it { should exist }
      its('content') { should match(/^DefaultEnvironment="LD_PRELOAD=#{LIBOTELINJECT_PATH}"$/) }
    else
      it { should_not exist }
    end
  end

  describe service('splunk-otel-collector') do
    it { should be_enabled }
    it { should be_running }
  end
end
