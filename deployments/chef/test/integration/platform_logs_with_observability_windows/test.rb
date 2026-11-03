splunk_access_token = 'testing123'
splunk_platform_url = 'https://fake-splunk-hec.com:8088/services/collector'
splunk_platform_token = 'fake-platform-token'
splunk_platform_logs_index = 'chef-test'
config_dir = "#{ENV['ProgramData']}\\Splunk\\OpenTelemetry Collector"
agent_config_path = "#{config_dir}\\agent_config.yaml"
logs_config_path = "#{config_dir}\\splunk_logs_config_windows.yaml"
service_key = 'HKLM:\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector'

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

describe file(logs_config_path) do
  it { should exist }
end

describe powershell("(Get-ItemProperty -Path '#{service_key}').Environment -join \"`n\"") do
  its('stdout') { should include("SPLUNK_ACCESS_TOKEN=#{splunk_access_token}") }
  its('stdout') { should include("SPLUNK_PLATFORM_URL=#{splunk_platform_url}") }
  its('stdout') { should include("SPLUNK_PLATFORM_TOKEN=#{splunk_platform_token}") }
  its('stdout') { should include("SPLUNK_PLATFORM_LOGS_INDEX=#{splunk_platform_logs_index}") }
end

describe registry_key('HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector') do
  its('ImagePath') do
    should match(
      /^.*--config "#{Regexp.escape(agent_config_path)}" --config "#{Regexp.escape(logs_config_path)}" --feature-gates=confmap\.enableMergeAppendOption$/i
    )
  end
end
