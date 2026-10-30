require 'spec_helper'

describe 'splunk_otel_collector' do
  let(:title) { 'splunk_otel_collector' }
  let(:params) { { 'splunk_access_token' => '' } }

  it "fails without an access token or Platform HEC token" do
    is_expected.to compile.and_raise_error(/splunk_access_token or splunk_platform_token/)
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
  windows_facts = on_supported_os.find { |os, _facts| os.start_with?('windows') }.last
  context 'with auto-instrumentation on Debian' do
    let(:facts) do
      on_supported_os.fetch('debian-12-x86_64').merge(
        'service_provider' => 'systemd',
        'local_groups' => '',
        'local_users' => '',
      )
    end
    let(:params) do
      {
        'splunk_access_token' => 'testing',
        'splunk_realm' => 'test',
        'with_auto_instrumentation' => true,
        'auto_instrumentation_version' => 'latest',
      }
    end

    it do
      is_expected.to contain_file('/etc/opentelemetry')
        .with_ensure('directory')
        .with_require('Package[splunk-otel-auto-instrumentation]')
    end
  end

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

  context 'with Splunk Platform logs on Linux' do
    let(:facts) do
      linux_facts.merge(service_provider: 'systemd', local_groups: '', local_users: '')
    end
    let(:params) do
      {
        'splunk_access_token' => 'testing',
        'splunk_realm' => 'test',
        'splunk_platform_token' => 'platform-token',
        'splunk_platform_url' => 'https://splunk.example.com:8088/services/collector',
        'splunk_platform_logs_index' => 'os-logs',
        'collector_command_line_args' => '--discovery',
      }
    end

    it do
      is_expected.to contain_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(%r{(?=.*^SPLUNK_ACCESS_TOKEN=testing$)(?=.*^SPLUNK_REALM=test$)(?=.*^SPLUNK_API_URL=https://api\.test\.observability\.splunkcloud\.com$)(?=.*^SPLUNK_HEC_TOKEN=testing$)(?=.*^SPLUNK_HEC_URL=https://ingest\.test\.observability\.splunkcloud\.com/v1/log$)(?=.*^SPLUNK_PLATFORM_URL=https://splunk\.example\.com:8088/services/collector$)(?=.*^SPLUNK_PLATFORM_TOKEN=platform-token$)(?=.*^SPLUNK_PLATFORM_LOGS_INDEX=os-logs$)(?=.*^SPLUNK_FILE_STORAGE_EXTENSION_PATH=/var/lib/otelcol/filelogs$)(?=.*^OTELCOL_OPTIONS=--discovery --config /etc/otel/collector/agent_config\.yaml --config /etc/otel/collector/splunk_logs_config_linux\.yaml --feature-gates=confmap\.enableMergeAppendOption$)(?!.*^SPLUNK_CONFIG=)}m)
    end

    it do
      is_expected.to contain_file('/var/lib/otelcol/filelogs').with(
        ensure: 'directory',
        owner: 'splunk-otel-collector',
        group: 'splunk-otel-collector',
        mode: '0700',
      ).that_requires('Package[splunk-otel-collector]')
        .that_requires('Class[splunk_otel_collector::collector_service_owner]')
        .that_comes_before('Service[splunk-otel-collector]')
    end
  end

  context 'with Splunk Platform logs on Windows' do
    let(:facts) do
      windows_facts.merge(
        'win_temp' => 'C:\\Windows\\Temp',
        'win_programfiles' => 'C:\\Program Files',
        'win_programdata' => 'C:\\ProgramData',
        'win_systemdrive' => 'C:',
      )
    end
    let(:params) do
      {
        'splunk_access_token' => 'testing',
        'splunk_realm' => 'test',
        'splunk_platform_token' => 'platform-token',
        'splunk_platform_url' => 'https://splunk.example.com:8088/services/collector',
        'splunk_platform_logs_index' => 'os-logs',
      }
    end

    it 'passes Platform settings to the MSI so it can compose the base and Windows logs configs' do
      install_options = catalogue.resource('Package', 'Splunk OpenTelemetry Collector')[:install_options]

      expect(install_options).to include(
        'SPLUNK_CONFIG' => 'C:\\ProgramData\\Splunk\\OpenTelemetry Collector\\agent_config.yaml',
        'SPLUNK_PLATFORM_URL' => 'https://splunk.example.com:8088/services/collector',
        'SPLUNK_PLATFORM_TOKEN' => 'platform-token',
        'SPLUNK_PLATFORM_LOGS_INDEX' => 'os-logs',
      )
    end

    it do
      is_expected.to contain_registry_value('HKLM\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector\\Environment')
        .with_data(array_including(
          'SPLUNK_PLATFORM_URL=https://splunk.example.com:8088/services/collector',
          'SPLUNK_PLATFORM_TOKEN=platform-token',
          'SPLUNK_PLATFORM_LOGS_INDEX=os-logs',
        ))
    end
  end

  context 'with Splunk Platform logs only on Windows' do
    let(:facts) do
      windows_facts.merge(
        'win_temp' => 'C:\\Windows\\Temp',
        'win_programfiles' => 'C:\\Program Files',
        'win_programdata' => 'C:\\ProgramData',
        'win_systemdrive' => 'C:',
      )
    end
    let(:params) do
      {
        'splunk_platform_token' => 'platform-token',
        'splunk_platform_url' => 'https://splunk.example.com:8088/services/collector',
      }
    end

    it { is_expected.to compile.with_all_deps }

    it do
      is_expected.to contain_registry_value('HKLM\\SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector\\Environment')
        .with_data(array_including(
          'SPLUNK_PLATFORM_URL=https://splunk.example.com:8088/services/collector',
          'SPLUNK_PLATFORM_TOKEN=platform-token',
        ))
    end
  end

  context 'with Splunk Platform logs only on Linux' do
    let(:facts) do
      linux_facts.merge(service_provider: 'systemd', local_groups: '', local_users: '')
    end
    let(:params) do
      {
        'splunk_platform_token' => 'platform-token',
        'splunk_platform_url' => 'https://splunk.example.com:8088/services/collector',
        'collector_command_line_args' => '--discovery',
      }
    end

    it do
      is_expected.to contain_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(%r{(?=.*^SPLUNK_PLATFORM_URL=https://splunk\.example\.com:8088/services/collector$)(?=.*^SPLUNK_PLATFORM_TOKEN=platform-token$)(?=.*^SPLUNK_FILE_STORAGE_EXTENSION_PATH=/var/lib/otelcol/filelogs$)(?=.*^OTELCOL_OPTIONS=--discovery --config /etc/otel/collector/splunk_logs_config_linux\.yaml --feature-gates=confmap\.enableMergeAppendOption$)(?!.*^SPLUNK_ACCESS_TOKEN=)(?!.*^SPLUNK_API_URL=)(?!.*^SPLUNK_CONFIG=)(?!.*^SPLUNK_HEC_TOKEN=)(?!.*^SPLUNK_HEC_URL=)(?!.*^SPLUNK_INGEST_URL=)(?!.*^SPLUNK_REALM=)}m)
    end
  end

  context 'with Splunk Platform logs and no explicit index' do
    let(:facts) do
      linux_facts.merge(service_provider: 'systemd', local_groups: '', local_users: '')
    end
    let(:params) do
      {
        'splunk_access_token' => 'testing',
        'splunk_realm' => 'test',
        'splunk_platform_token' => 'platform-token',
        'splunk_platform_url' => 'https://splunk.example.com:8088/services/collector',
      }
    end

    it do
      is_expected.to contain_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(%r{\A(?=.*^SPLUNK_PLATFORM_TOKEN=platform-token$)(?!.*^SPLUNK_PLATFORM_LOGS_INDEX=).*\z}m)
    end
  end

  context 'with Splunk Platform parameters missing the required URL or token' do
    let(:facts) do
      linux_facts.merge(service_provider: 'systemd', local_groups: '', local_users: '')
    end

    context 'when only the token is set' do
      let(:params) do
        { 'splunk_access_token' => 'testing', 'splunk_realm' => 'test', 'splunk_platform_token' => 'platform-token' }
      end

      it { is_expected.to compile.and_raise_error(/splunk_platform_url parameter is required/) }
    end

    context 'when only the URL is set' do
      let(:params) do
        { 'splunk_access_token' => 'testing', 'splunk_realm' => 'test', 'splunk_platform_url' => 'https://splunk.example.com' }
      end

      it { is_expected.to compile.and_raise_error(/splunk_platform_token parameter is required/) }
    end

    context 'when only the logs index is set' do
      let(:params) do
        { 'splunk_access_token' => 'testing', 'splunk_realm' => 'test', 'splunk_platform_logs_index' => 'os-logs' }
      end

      it { is_expected.to compile.and_raise_error(/splunk_platform_url parameter is required/) }
    end
  end
end
