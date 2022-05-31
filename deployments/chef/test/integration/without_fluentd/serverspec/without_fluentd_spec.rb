require 'spec_helper'

describe service('splunk-otel-collector') do
  it { should be_enabled }
  it { should be_running }
end

if os[:family] == 'windows'
  describe service('fluentdwinsvc') do
    it { should_not be_enabled }
    it { should_not be_running }
  end
else
  describe service('td-agent') do
    it { should_not be_enabled }
    it { should_not be_running }
  end
end
