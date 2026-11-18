splunk_access_token = 'testing123'
splunk_platform_url = 'https://fake-splunk-hec.com:8088/services/collector'
splunk_platform_token = 'fake-platform-token'
splunk_platform_logs_index = 'chef-test'

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

describe file('/etc/otel/collector/splunk-otel-collector.conf') do
  its('content') { should match /^SPLUNK_ACCESS_TOKEN=#{splunk_access_token}$/ }
  its('content') { should match /^SPLUNK_PLATFORM_URL=#{Regexp.escape(splunk_platform_url)}$/ }
  its('content') { should match /^SPLUNK_PLATFORM_TOKEN=#{splunk_platform_token}$/ }
  its('content') { should match /^SPLUNK_PLATFORM_LOGS_INDEX=#{splunk_platform_logs_index}$/ }
  its('content') { should match %r{^SPLUNK_FILE_STORAGE_EXTENSION_PATH=/var/lib/otelcol/filelogs$} }
  its('content') do
    should match %r{^OTELCOL_OPTIONS=--config /etc/otel/collector/agent_config\.yaml --config /etc/otel/collector/splunk_logs_config_linux\.yaml --feature-gates=confmap\.enableMergeAppendOption$}
  end
  its('content') { should_not match /^SPLUNK_CONFIG=/ }
end

describe file('/var/lib/otelcol/filelogs') do
  it { should be_directory }
  its('owner') { should eq 'splunk-otel-collector' }
  its('group') { should eq 'splunk-otel-collector' }
  its('mode') { should cmp '0700' }
end
