# Cookbook:: splunk_otel_collector
# Recipe:: collector_deb_repo

remote_file '/etc/apt/trusted.gpg.d/splunk.gpg' do
  source node['splunk_otel_collector']['debian_gpg_key_url']
  mode '0644'
  action :create
end

if node['splunk_otel_collector']['local_artifact_testing_enabled']

  file_name = 'soc.deb'
  deb_install_dir = '/etc/otel/collector/'
  deb_install_path = deb_install_dir + file_name

  # Create destination dir on target machine
  directory deb_install_dir do
    mode '0644'
    action :create
    recursive true
  end

  # Copy deb file from source to target
  cookbook_file deb_install_path do
    source file_name
    mode '0644'
  end

  dpkg_package deb_install_path do
    source deb_install_path
    action :install
  end

else
  file '/etc/apt/sources.list.d/splunk-otel-collector.list' do
    content "deb #{node['splunk_otel_collector']['debian_repo_url']} #{node['splunk_otel_collector']['package_stage']} main\n"
    mode '0644'
    notifies :update, 'apt_update[update apt cache]', :immediately
  end
end

apt_update 'update apt cache' do
  action :nothing
end
