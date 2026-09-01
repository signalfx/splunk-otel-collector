#
# Cookbook:: signalfx_agent
# Spec:: default
#
# Copyright:: 2018, The Authors, All Rights Reserved.

require 'spec_helper'

describe 'splunk_otel_collector::default' do
  context 'on the Linux platform family' do
    before do
      stub_command(
        'find /etc/otel/collector /var/lib/otelcol \( ! -user splunk-otel-collector ' \
        '-o ! -group splunk-otel-collector \) -print -quit | grep -q .'
      ).and_return(false)
    end

    context 'on debian-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
          }
        end.converge described_recipe
      end

      it 'converges successfully' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        expect { chef_run }.to_not raise_error
      end

      it_behaves_like 'common linux resources'
    end

    context 'on amazon-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'amazon', version: '2') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
          }
        end.converge described_recipe
      end

      it 'converges successfully' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        expect { chef_run }.to_not raise_error
      end

      it_behaves_like 'common linux resources'
    end

    context 'on RedHat-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'centos', version: '7') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
          }
        end.converge described_recipe
      end

      it 'converges successfully' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        expect { chef_run }.to_not raise_error
      end

      it_behaves_like 'common linux resources'
    end

    context 'on suse-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'suse', version: '15') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
          }
        end.converge described_recipe
      end

      it 'converges successfully' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        expect { chef_run }.to_not raise_error
      end

      it_behaves_like 'collector conf'
      it_behaves_like 'splunk-otel-collector linux service status'
      it_behaves_like 'install splunk-otel-collector package'
    end

    context 'with the last libsplunk auto-instrumentation release' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
            'with_auto_instrumentation' => true,
            'auto_instrumentation_version' => '0.159.0',
            'with_auto_instrumentation_sdks' => %w(java),
          }
        end.converge described_recipe
      end

      it 'manages the legacy zeroconfig layout' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        stub_command("bash -c 'command -v npm'").and_return(false)
        expect(chef_run).to create_template('/etc/splunk/zeroconfig/java.conf')
        expect(chef_run).to_not create_template('/etc/opentelemetry/injector/injector.conf')
        expect(chef_run).to delete_file('/etc/opentelemetry/injector/injector.conf')
      end
    end

    context 'with an OpenTelemetry injector auto-instrumentation release' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
            'with_auto_instrumentation' => true,
            'auto_instrumentation_version' => '0.159.0',
            'with_auto_instrumentation_sdks' => %w(java),
          }
        end.converge described_recipe
      end

      it 'manages the OpenTelemetry injector layout' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        stub_command("bash -c 'command -v npm'").and_return(false)
        expect(chef_run).to create_template('/etc/opentelemetry/injector/injector.conf')
        expect(chef_run).to create_template('/etc/opentelemetry/injector/default_env.conf')
        expect(chef_run).to delete_file('/etc/splunk/zeroconfig/java.conf')
      end
    end

    context 'with local auto-instrumentation artifact on debian-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '22.04') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
            'local_artifact_testing_enabled' => true,
            'with_auto_instrumentation' => true,
            'with_auto_instrumentation_sdks' => %w(java),
          }
        end.converge described_recipe
      end

      it 'installs the local auto-instrumentation deb package' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        stub_command("bash -c 'command -v npm'").and_return(false)
        expect(chef_run).to create_cookbook_file('/tmp/soai.deb').with(source: 'soai.deb', mode: '0644')
        expect(chef_run).to install_dpkg_package('splunk-otel-auto-instrumentation').with(source: '/tmp/soai.deb')
      end
    end

    context 'with local auto-instrumentation artifact on RedHat-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'centos', version: '7') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
            'local_artifact_testing_enabled' => true,
            'with_auto_instrumentation' => true,
            'with_auto_instrumentation_sdks' => %w(java),
          }
        end.converge described_recipe
      end

      it 'installs the local auto-instrumentation rpm package' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        stub_command("bash -c 'command -v npm'").and_return(false)
        expect(chef_run).to create_cookbook_file('/tmp/soai.rpm').with(source: 'soai.rpm', mode: '0644')
        expect(chef_run).to install_rpm_package('splunk-otel-auto-instrumentation').with(source: '/tmp/soai.rpm')
      end
    end

    context 'with local auto-instrumentation artifact on suse-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'suse', version: '15') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
            'local_artifact_testing_enabled' => true,
            'with_auto_instrumentation' => true,
            'with_auto_instrumentation_sdks' => %w(java),
          }
        end.converge described_recipe
      end

      it 'installs the local auto-instrumentation rpm package with zypper' do
        stub_command('getent group splunk-otel-collector').and_return(true)
        stub_command('getent passwd splunk-otel-collector').and_return(true)
        stub_command("bash -c 'command -v npm'").and_return(false)
        expect(chef_run).to create_cookbook_file('/tmp/soai.rpm').with(source: 'soai.rpm', mode: '0644')
        expect(chef_run).to install_zypper_package('splunk-otel-auto-instrumentation').with(source: '/tmp/soai.rpm', gpg_check: false)
      end
    end
  end
  context 'on the Windows platform family' do
    context 'on windows-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'windows', version: '2019') do |node|
          node.normal['splunk_otel_collector'] = {
            'splunk_access_token' => 'test123',
            'splunk_realm' => 'test',
            'collector_version' => '0.41.1',
          }
        end.converge described_recipe
      end

      it 'converges successfully' do
        expect { chef_run }.to_not raise_error
      end

      it_behaves_like 'common windows resources'
    end
  end
end
