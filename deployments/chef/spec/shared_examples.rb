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

shared_examples_for 'td-agent linux service status' do
  it 'enables td-agent service on startup' do
    expect(chef_run).to enable_service('td-agent')
  end

  it 'restart td-agent service on config change' do
    expect(chef_run.package('td-agent')).to notify('service[td-agent]').delayed
  end

  it 'starts the td-agent service' do
    expect(chef_run).to start_service 'td-agent'
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

shared_examples_for 'td-agent windows service status' do
  it 'enables td-agent service on startup' do
    expect(chef_run).to enable_windows_service('td-agent')
  end

  it 'starts the td-agent service' do
    expect(chef_run).to start_windows_service 'fluentdwinsvc'
  end
end

shared_examples_for 'collector conf' do
  it 'does not complain about a missing splunk_access_token' do
    expect(chef_run).not_to run_ruby_block('splunk-access-token-unset')
  end

  it 'does not complain about a missing splunk_realm' do
    expect(chef_run).not_to run_ruby_block('splunk-realm-unset')
  end

  it 'does not complain about a missing splunk_access_token' do
    expect(chef_run).not_to run_ruby_block('splunk-access-token-unset')
  end

  it 'does not complain about a missing splunk_realm' do
    expect(chef_run).not_to run_ruby_block('splunk-realm-unset')
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

shared_examples_for 'install td-agent package' do
  it 'installs td-agent package' do
    expect(chef_run).to install_package('td-agent')
  end
end

shared_examples_for 'common linux resources' do
  it_behaves_like 'collector conf'
  it_behaves_like 'splunk-otel-collector linux service status'
  it_behaves_like 'td-agent linux service status'
  it_behaves_like 'install splunk-otel-collector package'
  it_behaves_like 'install td-agent package'
end

shared_examples_for 'common windows resources' do
  it_behaves_like 'collector conf'
  it_behaves_like 'splunk-otel-collector windows service status'
  it_behaves_like 'td-agent windows service status'

  it 'installs splunk-otel-collector package' do
    expect(chef_run).to install_windows_package('splunk-otel-collector')
  end

  it 'installs td-agent package' do
    expect(chef_run).to install_windows_package('fluentd')
  end
end
