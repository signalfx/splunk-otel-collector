require 'spec_helper'

describe 'splunk_otel_collector' do
  let(:title) { 'splunk_otel_collector' }
  let(:params) { { 'splunk_access_token' => '' } }

  it "fails without access token" do
    is_expected.to compile.and_raise_error(/splunk_access_token/)
  end

  on_supported_os.each do |os, facts|
    # When running 'rake spec' it checks for absolute paths and on Windows paths
    # are built from $facts. In order to pass the tests the required $facts are
    # being explicitly added.
    let(:facts) {{
      'win_temp' =>'C:\\Windows\\Temp',
      'win_programfiles' => 'C:\\Program Files',
      'win_programdata' => 'C:\\ProgramData',
      'win_systemdrive' => 'C:'}.merge( facts )
    }
    context "on #{os}" do
      let(:params) { { 'splunk_access_token' => "testing", 'splunk_realm' => 'test' } }
      it { is_expected.to compile.with_all_deps }
    end
  end
end
