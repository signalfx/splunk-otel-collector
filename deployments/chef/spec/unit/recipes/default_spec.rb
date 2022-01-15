#
# Cookbook:: signalfx_agent
# Spec:: default
#
# Copyright:: 2018, The Authors, All Rights Reserved.

require 'spec_helper'

describe 'splunk-otel-collector::default' do
  context 'on the Linux platform family' do
    context 'on debian-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'ubuntu', version: '16.04') do |node|
          node.normal['splunk-otel-collector'] = {
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
          node.normal['splunk-otel-collector'] = {
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
          node.normal['splunk-otel-collector'] = {
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
          node.normal['splunk-otel-collector'] = {
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

      it 'enables td-agent service on startup' do
        expect(chef_run).not_to enable_service('td-agent')
      end

      it 'restart td-agent service on config change' do
        expect(chef_run.package('td-agent')).not_to notify('service[td-agent]').delayed
      end

      it 'starts the td-agent service' do
        expect(chef_run).not_to start_service 'td-agent'
      end

      it 'installs td-agent package' do
        expect(chef_run).not_to install_package('td-agent')
      end
    end
  end
  context 'on the Windows platform family' do
    context 'on windows-family distro' do
      cached(:chef_run) do
        ChefSpec::SoloRunner.new(platform: 'windows', version: '2019') do |node|
          node.normal['splunk-otel-collector'] = {
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
