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

  linux_facts = on_supported_os.find { |os, _facts| !os.start_with?('windows') }.last
  context 'with Linux service owner resources' do
    let(:facts) do
      linux_facts.merge(service_provider: 'systemd', local_groups: '', local_users: '')
    end
    let(:params) do
      {
        'splunk_access_token' => 'testing',
        'splunk_realm' => 'test',
        'service_user' => 'custom-user',
        'service_group' => 'custom-group',
      }
    end

    ['/etc/otel/collector', '/var/lib/otelcol'].each do |path|
      it do
        is_expected.to contain_file(path).with(
          ensure: 'directory',
          owner: 'custom-user',
          group: 'custom-group',
          mode: '0755',
        ).that_requires('Exec[systemctl stop splunk-otel-collector]')
          .that_comes_before('Exec[set collector directory ownership recursively]')
      end
    end

    it do
      is_expected.to contain_exec('set collector directory ownership recursively').with(
        command: 'chown -R custom-user:custom-group /etc/otel/collector /var/lib/otelcol',
        onlyif: 'find /etc/otel/collector /var/lib/otelcol ' \
                '\( ! -user custom-user -o ! -group custom-group \) -print -quit | grep -q .',
      ).that_comes_before('Exec[systemd-tmpfiles --create --remove /etc/tmpfiles.d/splunk-otel-collector.conf]')
    end

    it do
      is_expected.to contain_exec('systemctl stop splunk-otel-collector')
        .with(refreshonly: true)
        .that_subscribes_to('File_line[/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf]')
        .that_subscribes_to('File_line[set-service-user]')
        .that_subscribes_to('File_line[set-service-group]')
    end

    it do
      is_expected.to contain_exec('systemctl daemon-reload')
        .that_subscribes_to('File_line[/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf]')
        .that_requires('Exec[systemd-tmpfiles --create --remove /etc/tmpfiles.d/splunk-otel-collector.conf]')
        .that_notifies('Service[splunk-otel-collector]')
    end
  end
end
