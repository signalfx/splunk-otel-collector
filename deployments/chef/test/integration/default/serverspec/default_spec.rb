require 'spec_helper'

splunk_access_token = 'testing123'
splunk_realm = 'test'
splunk_api_url = "https://api.#{splunk_realm}.signalfx.com"
splunk_ingest_url = "https://ingest.#{splunk_realm}.signalfx.com"
splunk_trace_url = "#{splunk_ingest_url}/v2/trace"
splunk_hec_url = "#{splunk_ingest_url}/v1/log"
splunk_hec_token = splunk_access_token
splunk_memory_total = '512'

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

if os[:family] == 'windows'
  bundle_dir = 'C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle'
  collectd_dir = "#{bundle_dir}\\run\\collectd"
  config_path = 'C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml'
  describe windows_registry_key('HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment') do
    it { should have_property_value('SPLUNK_ACCESS_TOKEN', :type_string, splunk_access_token) }
    it { should have_property_value('SPLUNK_API_URL', :type_string, splunk_api_url) }
    it { should have_property_value('SPLUNK_BUNDLE_DIR', :type_string, bundle_dir) }
    it { should have_property_value('SPLUNK_COLLECTD_DIR', :type_string, collectd_dir) }
    it { should have_property_value('SPLUNK_CONFIG', :type_string, config_path) }
    it { should have_property_value('SPLUNK_HEC_TOKEN', :type_string, splunk_hec_token) }
    it { should have_property_value('SPLUNK_HEC_URL', :type_string, splunk_hec_url) }
    it { should have_property_value('SPLUNK_INGEST_URL', :type_string, splunk_ingest_url) }
    it { should have_property_value('SPLUNK_MEMORY_TOTAL_MIB', :type_string, splunk_memory_total) }
    it { should have_property_value('SPLUNK_REALM', :type_string, splunk_realm) }
    it { should have_property_value('SPLUNK_TRACE_URL', :type_string, splunk_trace_url) }
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
    its('content') { should match /^SPLUNK_MEMORY_TOTAL_MIB=#{splunk_memory_total}$/ }
    its('content') { should match /^SPLUNK_REALM=test$/ }
    its('content') { should match /^SPLUNK_TRACE_URL=#{splunk_trace_url}$/ }
  end
  describe file('/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf') do
    its('content') { should match /^User=splunk-otel-collector$/ }
    its('content') { should match /^Group=splunk-otel-collector$/ }
  end
  if os[:family] != 'suse' && os[:family] != 'opensuse'
    if os[:family] != 'ubuntu' || os[:release] != '22.04'
      fluentd_config_path = '/etc/otel/collector/fluentd/fluent.conf'
      describe service('td-agent') do
        it { should be_enabled }
        it { should be_running }
      end
      describe file('/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf') do
        its('content') { should match /^Environment=FLUENT_CONF=#{fluentd_config_path}$/ }
      end
    end
  end
  describe package('splunk-otel-auto-instrumentation') do
    it { should_not be_installed }
  end
end
