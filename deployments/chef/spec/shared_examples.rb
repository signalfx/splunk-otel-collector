shared_examples_for 'splunk-otel-collector linux service status' do
  it 'enables splunk-otel-collector service on startup' do
    expect(chef_run).to enable_service('splunk-otel-collector')
  end

  it 'restart splunk-otel-collector service on config change' do
    expect(chef_run.package('splunk-otel-collector')).to notify('service[splunk-otel-collector]').delayed
  end

  it 'starts the splunk-otel-collector service' do
    expect(chef_run).to start_service 'splunk-otel-collector'
  end
end

shared_examples_for 'splunk-otel-collector windows service status' do
  it 'enables splunk-otel-collector service on startup' do
    expect(chef_run).to enable_windows_service('splunk-otel-collector')
  end

  it 'starts the splunk-otel-collector service' do
    expect(chef_run).to start_windows_service 'splunk-otel-collector'
  end
end

shared_examples_for 'collector conf' do
  it 'does not complain about a missing splunk_access_token' do
    expect(chef_run).not_to run_ruby_block('splunk-access-token-unset')
  end
end

shared_examples_for 'install splunk-otel-collector package' do
  it 'installs splunk-otel-collector package' do
    expect(chef_run).to install_package('splunk-otel-collector')
  end

  it 'drops an agent config file' do
    expect(chef_run).to create_template '/etc/otel/collector/splunk-otel-collector.conf'
  end
end

shared_examples_for 'common linux resources' do
  it_behaves_like 'collector conf'
  it_behaves_like 'splunk-otel-collector linux service status'
  it_behaves_like 'install splunk-otel-collector package'

  %w(/etc/otel/collector /var/lib/otelcol).each do |path|
    it "manages #{path} for the collector service owner" do
      expect(chef_run).to create_directory("ensure #{path} exists").with(
        path: path,
        mode: '0755'
      )
    end
  end

  it 'sets collector directory ownership recursively' do
    expect(chef_run.execute('set collector directory ownership recursively').command).to eq(
      'chown -R -- splunk-otel-collector:splunk-otel-collector /etc/otel/collector /var/lib/otelcol'
    )
  end

  it 'stops the collector immediately when the service owner configuration changes' do
    owner_config = chef_run.file('/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf')
    expect(owner_config).to notify('service[splunk-otel-collector]').to(:stop).immediately
  end
end

shared_examples_for 'common windows resources' do
  it_behaves_like 'collector conf'
  it_behaves_like 'splunk-otel-collector windows service status'

  it 'installs splunk-otel-collector package' do
    expect(chef_run).to install_windows_package('splunk-otel-collector')
  end
end
