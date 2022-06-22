require 'spec_helper'

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
end

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

if os[:family] != 'suse' && os[:family] != 'opensuse'
  if os[:family] != 'ubuntu' || os[:release] != '22.04'
    describe service('td-agent') do
      it { should be_enabled }
      it { should be_running }
    end
  end
end
