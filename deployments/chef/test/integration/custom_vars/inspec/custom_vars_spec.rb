splunk_access_token = 'testing123'
splunk_realm = 'test'
splunk_api_url = 'https://fake-splunk-api.com'
splunk_ingest_url = 'https://fake-splunk-ingest.com'
splunk_trace_url = "#{splunk_ingest_url}/v2/trace"
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
  describe registry_key('HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment') do
    its('SPLUNK_ACCESS_TOKEN') { should eq splunk_access_token }
    its('SPLUNK_API_URL') { should eq splunk_api_url }
    its('SPLUNK_BUNDLE_DIR') { should eq bundle_dir }
    its('SPLUNK_COLLECTD_DIR') { should eq collectd_dir }
    its('SPLUNK_CONFIG') { should eq config_path }
    its('SPLUNK_HEC_TOKEN') { should eq splunk_hec_token }
    its('SPLUNK_HEC_URL') { should eq splunk_hec_url }
    its('SPLUNK_INGEST_URL') { should eq splunk_ingest_url }
    its('SPLUNK_LISTEN_INTERFACE') { should eq splunk_listen_interface }
    its('SPLUNK_MEMORY_TOTAL_MIB') { should eq splunk_memory_total }
    its('SPLUNK_REALM') { should eq splunk_realm }
    its('SPLUNK_TRACE_URL') { should eq splunk_trace_url }
    its('MY_CUSTOM_VAR1') { should eq 'value1' }
    its('MY_CUSTOM_VAR2') { should eq 'value2' }
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
    its('content') { should match /^SPLUNK_TRACE_URL=#{splunk_trace_url}$/ }
    its('content') { should match /^MY_CUSTOM_VAR1=value1$/ }
    its('content') { should match /^MY_CUSTOM_VAR2=value2$/ }
  end
  describe file('/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf') do
    its('content') { should match /^User=custom-user$/ }
    its('content') { should match /^Group=custom-group$/ }
  end
  if os[:family] != 'suse' && os[:family] != 'opensuse'
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
