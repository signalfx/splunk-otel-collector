# Cookbook:: splunk_otel_collector
# Recipe:: auto_instrumentation

package 'splunk-otel-auto-instrumentation' do
  action :install
  version node['splunk_otel_collector']['auto_instrumentation_version'] if node['splunk_otel_collector']['auto_instrumentation_version'] != 'latest'
  flush_cache [ :before ] if platform_family?('amazon', 'rhel')
  options '--allow-downgrades' if platform_family?('debian') \
    && node['packages'] \
    && node['packages']['apt'] \
    && Gem::Version.new(node['packages']['apt']['version'].split('~')[0]) >= Gem::Version.new('1.1.0')
  allow_downgrade true if platform_family?('amazon', 'rhel', 'suse')
end

template '/etc/ld.so.preload' do
  source 'ld.so.preload.erb'
end

template '/usr/lib/splunk-instrumentation/instrumentation.conf' do
  source 'instrumentation.conf.erb'
end
