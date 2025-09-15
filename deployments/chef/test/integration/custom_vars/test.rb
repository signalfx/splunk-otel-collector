splunk_access_token = 'testing123'
splunk_realm = 'test'
splunk_api_url = 'https://fake-splunk-api.com'
splunk_ingest_url = 'https://fake-splunk-ingest.com'
splunk_hec_url = "#{splunk_ingest_url}/v1/log"
splunk_hec_token = 'fake-hec-token'
splunk_memory_total = '256'
splunk_listen_interface = '0.0.0.0'

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

if os[:family] == 'windows'
  bundle_dir = "#{ENV['ProgramFiles']}\\Splunk\\OpenTelemetry Collector\\agent-bundle"
  collectd_dir = "#{bundle_dir}\\run\\collectd"
  config_path = "#{ENV['ProgramData']}\\Splunk\\OpenTelemetry Collector\\agent_config.yaml"
  collector_env_vars = [
    { name: 'SPLUNK_ACCESS_TOKEN', type: :string, data: splunk_access_token },
    { name: 'SPLUNK_API_URL', type: :string, data: splunk_api_url },
    { name: 'SPLUNK_BUNDLE_DIR', type: :string, data: bundle_dir },
    { name: 'SPLUNK_COLLECTD_DIR', type: :string, data: collectd_dir },
    { name: 'SPLUNK_CONFIG', type: :string, data: config_path },
    { name: 'SPLUNK_HEC_TOKEN', type: :string, data: splunk_hec_token },
    { name: 'SPLUNK_HEC_URL', type: :string, data: splunk_hec_url },
    { name: 'SPLUNK_INGEST_URL', type: :string, data: splunk_ingest_url },
    { name: 'SPLUNK_LISTEN_INTERFACE', type: :string, data: splunk_listen_interface },
    { name: 'SPLUNK_MEMORY_TOTAL_MIB', type: :string, data: splunk_memory_total },
    { name: 'SPLUNK_REALM', type: :string, data: splunk_realm },
    { name: 'MY_CUSTOM_VAR1', type: :string, data: 'value1' },
    { name: 'MY_CUSTOM_VAR2', type: :string, data: 'value2' },
  ]
  unless splunk_listen_interface.to_s.strip.empty?
    collector_env_vars.push({ name: 'SPLUNK_LISTEN_INTERFACE', type: :string, data: splunk_listen_interface })
  end
  collector_env_vars_strings = []
  collector_env_vars.each do |item|
    collector_env_vars_strings |= [ "#{item[:name]}=#{item[:data]}" ]
  end
  collector_env_vars_strings.sort!
  describe registry_key('HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\splunk-otel-collector') do
    it { should have_property 'Environment' }
    it { should have_property_value('Environment', :multi_string, collector_env_vars_strings) }
  end
  describe registry_key('HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\splunk-otel-collector') do
    it { should have_property 'ImagePath' }
    its('ImagePath') { should match /^.*--discovery --set=processors.batch.timeout=10s$/ }
  end
  describe service('fluentdwinsvc') do
    it { should be_enabled }
    it { should be_running }
  end
else
  bundle_dir = '/usr/lib/splunk-otel-collector/agent-bundle'
  collectd_dir = "#{bundle_dir}/run/collectd"
  config_path = '/etc/otel/collector/agent_config.yaml'
  describe file('/etc/otel/collector/splunk-otel-collector.conf') do
    its('content') { should match /^SPLUNK_ACCESS_TOKEN=#{splunk_access_token}$/ }
    its('content') { should match /^SPLUNK_API_URL=#{splunk_api_url}$/ }
    its('content') { should match /^SPLUNK_BUNDLE_DIR=#{bundle_dir}$/ }
    its('content') { should match /^SPLUNK_COLLECTD_DIR=#{collectd_dir}$/ }
    its('content') { should match /^SPLUNK_CONFIG=#{config_path}$/ }
    its('content') { should match /^SPLUNK_HEC_TOKEN=#{splunk_hec_token}$/ }
    its('content') { should match /^SPLUNK_HEC_URL=#{splunk_hec_url}$/ }
    its('content') { should match /^SPLUNK_INGEST_URL=#{splunk_ingest_url}$/ }
    its('content') { should match /^SPLUNK_LISTEN_INTERFACE=#{splunk_listen_interface}$/ }
    its('content') { should match /^SPLUNK_MEMORY_TOTAL_MIB=#{splunk_memory_total}$/ }
    its('content') { should match /^SPLUNK_REALM=test$/ }
    its('content') { should match /^MY_CUSTOM_VAR1=value1$/ }
    its('content') { should match /^MY_CUSTOM_VAR2=value2$/ }
    its('content') { should match /^OTELCOL_OPTIONS=--discovery --set=processors.batch.timeout=10s$/ }
  end
  describe file('/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf') do
    its('content') { should match /^User=custom-user$/ }
    its('content') { should match /^Group=custom-group$/ }
  end
  if os[:family] != 'suse' && os[:family] != 'opensuse' && !(os[:family] == 'debian' && ::Gem::Version.new(os.release) >= ::Gem::Version.new('11'))
    fluentd_config_path = '/etc/otel/collector/fluentd/fluent.conf'
    describe service('td-agent') do
      it { should be_enabled }
      it { should be_running }
    end
    describe file('/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf') do
      its('content') { should match /^Environment=FLUENT_CONF=#{fluentd_config_path}$/ }
    end
  end
  describe package('splunk-otel-auto-instrumentation') do
    it { should_not be_installed }
  end
end
