# frozen_string_literal: true

require 'spec_helper'

describe 'splunk_otel_collector::default' do
  before do
    stub_command(
      'find /etc/otel/collector /var/lib/otelcol \( ! -user splunk-otel-collector ' \
      '-o ! -group splunk-otel-collector \) -print -quit | grep -q .'
    ).and_return(false)
    stub_command('getent group splunk-otel-collector').and_return(true)
    stub_command('getent passwd splunk-otel-collector').and_return(true)
  end

  context 'with Splunk Platform HEC as the only destination' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_platform_url' => 'https://hec.example.com:8088',
          'splunk_platform_token' => 'platform-token',
          'splunk_platform_logs_index' => 'host-logs',
          'collector_command_line_args' => '--set=custom',
        }
      end.converge described_recipe
    end

    it 'does not require an Observability access token' do
      expect { chef_run }.not_to raise_error
    end

    it 'configures the Platform HEC environment without Observability settings' do
      expect(chef_run).to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(%r{^SPLUNK_PLATFORM_URL=https://hec\.example\.com:8088$})
        .with_content(/^SPLUNK_PLATFORM_TOKEN=platform-token$/)
        .with_content(/^SPLUNK_PLATFORM_LOGS_INDEX=host-logs$/)
        .with_content(
          %r{^SPLUNK_FILE_STORAGE_EXTENSION_PATH=/var/lib/otelcol/filelogs$}
        )
        .with_content(
          %r{^OTELCOL_OPTIONS=--set=custom --config /etc/otel/collector/splunk_logs_config_linux\.yaml$}
        )

      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_CONFIG=/)
      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_ACCESS_TOKEN=/)
      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_LISTEN_INTERFACE=/)
      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/confmap\.enableMergeAppendOption/)
    end

    it 'creates the file storage directory for the collector user' do
      expect(chef_run).to create_directory('/var/lib/otelcol/filelogs').with(
        owner: 'splunk-otel-collector',
        group: 'splunk-otel-collector',
        mode: '0700',
        recursive: true
      )
    end
  end

  context 'with Splunk Platform HEC and no explicit index' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_platform_url' => 'https://hec.example.com:8088',
          'splunk_platform_token' => 'platform-token',
        }
      end.converge described_recipe
    end

    it 'leaves the index unset so Splunk uses the HEC token default index' do
      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_PLATFORM_LOGS_INDEX=/)
    end
  end

  context 'with Splunk Platform HEC alongside Observability' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_access_token' => 'observability-token',
          'splunk_realm' => 'test',
          'splunk_platform_url' => 'https://hec.example.com:8088',
          'splunk_platform_token' => 'platform-token',
        }
      end.converge described_recipe
    end

    it 'composes both collector configs and enables the merge-append feature gate' do
      expect(chef_run).to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_ACCESS_TOKEN=observability-token$/)
        .with_content(%r{^SPLUNK_PLATFORM_URL=https://hec\.example\.com:8088$})
        .with_content(
          %r{^OTELCOL_OPTIONS=--config /etc/otel/collector/agent_config\.yaml --config /etc/otel/collector/splunk_logs_config_linux\.yaml --feature-gates=confmap\.enableMergeAppendOption$}
        )

      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_CONFIG=/)
    end
  end

  context 'with Observability and no Platform HEC configuration' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_access_token' => 'observability-token',
          'splunk_realm' => 'test',
        }
      end.converge described_recipe
    end

    it 'keeps the existing collector config environment variable' do
      expect(chef_run).to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(%r{^SPLUNK_CONFIG=/etc/otel/collector/agent_config\.yaml$})
        .with_content(/^SPLUNK_ACCESS_TOKEN=observability-token$/)
      expect(chef_run).not_to render_file('/etc/otel/collector/splunk-otel-collector.conf')
        .with_content(/^SPLUNK_PLATFORM_URL=/)
    end
  end

  context 'when Platform URL is set without its token' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_platform_url' => 'https://hec.example.com:8088',
        }
      end.converge described_recipe
    end

    it 'raises a validation error' do
      expect { chef_run }.to raise_error(RuntimeError, /splunk_platform_token/)
    end
  end

  context 'when Platform token is set without its URL' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_platform_token' => 'platform-token',
        }
      end.converge described_recipe
    end

    it 'raises a validation error' do
      expect { chef_run }.to raise_error(RuntimeError, /splunk_platform_url/)
    end
  end

  context 'when an index is set without the Platform URL' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
        node.normal['splunk_otel_collector'] = {
          'splunk_platform_logs_index' => 'host-logs',
        }
      end.converge described_recipe
    end

    it 'raises a validation error' do
      expect { chef_run }.to raise_error(RuntimeError, /splunk_platform_url/)
    end
  end

  context 'with Splunk Platform HEC as the only destination on Windows' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'windows', version: '2019') do |node|
        node.normal['splunk_otel_collector'].merge!(
          'collector_version' => '0.127.0',
          'splunk_platform_url' => 'https://hec.example.com:8088',
          'splunk_platform_token' => 'platform-token',
          'splunk_platform_logs_index' => 'host-logs'
        )
      end.converge described_recipe
    end

    it 'passes the HEC settings and lets the MSI select the packaged logs config' do
      expect { chef_run }.not_to raise_error
      env_vars = chef_run.node['splunk_otel_collector']['collector_win_env_vars']

      expect(env_vars).to include(
        name: 'SPLUNK_PLATFORM_URL', type: :string, data: 'https://hec.example.com:8088'
      )
      expect(env_vars).to include(
        name: 'SPLUNK_PLATFORM_TOKEN', type: :string, data: 'platform-token'
      )
      expect(env_vars).to include(
        name: 'SPLUNK_PLATFORM_LOGS_INDEX', type: :string, data: 'host-logs'
      )
      expect(env_vars).not_to include(
        hash_including(name: 'SPLUNK_CONFIG')
      )
    end
  end

  context 'with Splunk Platform HEC alongside Observability on Windows' do
    let(:chef_run) do
      ChefSpec::SoloRunner.new(platform: 'windows', version: '2019') do |node|
        node.normal['splunk_otel_collector'].merge!(
          'collector_version' => '0.127.0',
          'splunk_access_token' => 'observability-token',
          'splunk_platform_url' => 'https://hec.example.com:8088',
          'splunk_platform_token' => 'platform-token',
          'splunk_platform_logs_index' => 'host-logs',
          'collector_command_line_args' => '--discovery'
        )
      end.converge described_recipe
    end

    it 'passes the base config and Platform settings for MSI config composition' do
      expect { chef_run }.not_to raise_error
      env_vars = chef_run.node['splunk_otel_collector']['collector_win_env_vars']

      expect(env_vars).to include(
        name: 'SPLUNK_CONFIG',
        type: :string,
        data: chef_run.node['splunk_otel_collector']['collector_config_dest']
      )
      expect(env_vars).to include(
        name: 'SPLUNK_PLATFORM_URL', type: :string, data: 'https://hec.example.com:8088'
      )
      expect(env_vars).to include(
        name: 'SPLUNK_PLATFORM_TOKEN', type: :string, data: 'platform-token'
      )
      expect(env_vars).to include(
        name: 'SPLUNK_PLATFORM_LOGS_INDEX', type: :string, data: 'host-logs'
      )
      expect(chef_run.node['splunk_otel_collector']['collector_command_line_args']).to eq('--discovery')
    end
  end
end
