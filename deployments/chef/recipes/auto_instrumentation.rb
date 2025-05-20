# Cookbook:: splunk_otel_collector
# Recipe:: auto_instrumentation

with_new_instrumentation = node['splunk_otel_collector']['auto_instrumentation_version'] == 'latest' || Gem::Version.new(node['splunk_otel_collector']['auto_instrumentation_version']) >= Gem::Version.new('0.87.0')
with_systemd = node['splunk_otel_collector']['auto_instrumentation_systemd'].to_s.downcase == 'true'
with_java = node['splunk_otel_collector']['with_auto_instrumentation_sdks'].include?('java')
with_nodejs = node['splunk_otel_collector']['with_auto_instrumentation_sdks'].include?('nodejs') && with_new_instrumentation
dotnet_supported = %w(x86_64 amd64).include?(node['cpu']['architecture']) && (node['splunk_otel_collector']['auto_instrumentation_version'] == 'latest' || Gem::Version.new(node['splunk_otel_collector']['auto_instrumentation_version']) >= Gem::Version.new('0.99.0'))
with_dotnet = node['splunk_otel_collector']['with_auto_instrumentation_sdks'].include?('dotnet') && dotnet_supported
npm_path = node['splunk_otel_collector']['auto_instrumentation_npm_path']
lib_dir = '/usr/lib/splunk-instrumentation'
splunk_otel_js_path = "#{lib_dir}/splunk-otel-js.tgz"
splunk_otel_js_prefix = "#{lib_dir}/splunk-otel-js"
npm_install = "#{npm_path} install #{splunk_otel_js_path}"
zc_config_dir = '/etc/splunk/zeroconfig'
java_config_file = "#{zc_config_dir}/java.conf"
nodejs_config_file = "#{zc_config_dir}/node.conf"
dotnet_config_file = "#{zc_config_dir}/dotnet.conf"
old_config_file = "#{lib_dir}/instrumentation.conf"
systemd_config_dir = '/usr/lib/systemd/system.conf.d'
systemd_config_file = "#{systemd_config_dir}/00-splunk-otel-auto-instrumentation.conf"

# will be updated at run time based on whether npm is found
node.run_state[:with_nodejs] = with_nodejs && shell_out("bash -c 'command -v #{npm_path}'").exitstatus == 0

ohai 'reload packages' do
  action :nothing
  plugin 'packages'
end

execute 'reload systemd' do
  action :nothing
  command 'systemctl daemon-reload'
end

directory "#{splunk_otel_js_prefix}/node_modules" do
  action :nothing
  recursive true
end

execute npm_install do
  action :nothing
  cwd splunk_otel_js_prefix
end

ruby_block 'install splunk-otel-js' do
  action :nothing
  block do
    node.run_state[:with_nodejs] = shell_out("bash -c 'command -v #{npm_path}'").exitstatus == 0
    if node.run_state[:with_nodejs]
      resources(directory: "#{splunk_otel_js_prefix}/node_modules").run_action(:create)
      resources(execute: npm_install).run_action(:run)
    end
  end
  only_if { with_nodejs }
end

package 'splunk-otel-auto-instrumentation' do
  action :install
  version node['splunk_otel_collector']['auto_instrumentation_version'] if node['splunk_otel_collector']['auto_instrumentation_version'] != 'latest'
  flush_cache [ :before ] if platform_family?('amazon', 'rhel')
  options '--allow-downgrades' if platform_family?('debian') \
    && node['packages'] \
    && node['packages']['apt'] \
    && Gem::Version.new(node['packages']['apt']['version'].split('~').first) >= Gem::Version.new('1.1.0')
  allow_downgrade true if platform_family?('amazon', 'rhel', 'suse')
  notifies :reload, 'ohai[reload packages]', :immediately
  notifies :run, 'ruby_block[install splunk-otel-js]', :immediately
end

if with_systemd
  [java_config_file, nodejs_config_file, dotnet_config_file, old_config_file].each do |config_file|
    file config_file do
      action :delete
    end
  end
  directory systemd_config_dir do
    recursive true
  end
  template systemd_config_file do
    variables(
      installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] },
      with_java: lazy { with_java },
      with_nodejs: lazy { node.run_state[:with_nodejs] },
      with_dotnet: lazy { with_dotnet }
    )
    source '00-splunk-otel-auto-instrumentation.conf.erb'
    notifies :run, 'execute[reload systemd]', :delayed
    only_if { with_java || node.run_state[:with_nodejs] || with_dotnet }
  end
elsif with_new_instrumentation
  [old_config_file, systemd_config_file].each do |config_file|
    file config_file do
      action :delete
    end
  end
  file java_config_file do
    action :delete
    not_if { with_java }
  end
  file nodejs_config_file do
    action :delete
    not_if { node.run_state[:with_nodejs] }
  end
  file dotnet_config_file do
    action :delete
    not_if { with_dotnet }
  end
  directory zc_config_dir do
    recursive true
  end
  template java_config_file do
    variables(
      installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
    )
    source 'java.conf.erb'
    only_if { with_java }
  end
  template nodejs_config_file do
    variables(
      installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
    )
    source 'node.conf.erb'
    only_if { node.run_state[:with_nodejs] }
  end
  template dotnet_config_file do
    variables(
      installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
    )
    source 'dotnet.conf.erb'
    only_if { with_dotnet }
  end
else
  [java_config_file, nodejs_config_file, dotnet_config_file, systemd_config_file].each do |config_file|
    file config_file do
      action :delete
    end
  end
  template old_config_file do
    variables(
      installed_version: lazy { node['packages']['splunk-otel-auto-instrumentation']['version'] }
    )
    source 'instrumentation.conf.erb'
    only_if { with_java }
  end
end

template '/etc/ld.so.preload' do
  variables(
    with_systemd: lazy { with_systemd },
    with_java: lazy { with_java },
    with_nodejs: lazy { node.run_state[:with_nodejs] },
    with_dotnet: lazy { with_dotnet }
  )
  source 'ld.so.preload.erb'
end
