# Cookbook:: splunk_otel_collector
# Recipe:: auto_instrumentation

ohai 'reload packages' do
  plugin 'packages'
  action :nothing
end

package 'splunk-otel-auto-instrumentation' do
  action :install
  version node['splunk_otel_collector']['auto_instrumentation_version'] if node['splunk_otel_collector']['auto_instrumentation_version'] != 'latest'
  flush_cache [ :before ] if platform_family?('amazon', 'rhel')
  options '--allow-downgrades' if platform_family?('debian') \
    && node['packages'] \
    && node['packages']['apt'] \
    && Gem::Version.new(node['packages']['apt']['version'].split('~')[0]) >= Gem::Version.new('1.1.0')
  allow_downgrade true if platform_family?('amazon', 'rhel', 'suse')
  notifies :reload, 'ohai[reload packages]', :immediately
end

template '/etc/ld.so.preload' do
  source 'ld.so.preload.erb'
end

if node['splunk_otel_collector']['auto_instrumentation_systemd'].to_s.downcase == 'true'
  execute 'reload systemd' do
    command 'systemctl daemon-reload'
    action :nothing
  end
  directory '/usr/lib/systemd/system.conf.d' do
    recursive true
    action :create
  end
  template '/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf' do
    variables(
      installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
    )
    source '00-splunk-otel-auto-instrumentation.conf.erb'
    notifies :run, 'execute[reload systemd]', :immediately
  end
else
  file '/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf' do
    action :delete
  end
  if node['splunk_otel_collector']['auto_instrumentation_version'] == 'latest' || Gem::Version.new(node['splunk_otel_collector']['auto_instrumentation_version']) >= Gem::Version.new('0.87.0')
    template '/etc/splunk/zeroconfig/java.conf' do
      variables(
        installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
      )
      source 'java.conf.erb'
    end
  else
    template '/usr/lib/splunk-instrumentation/instrumentation.conf' do
      variables(
        installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
      )
      source 'instrumentation.conf.erb'
    end
  end
end
