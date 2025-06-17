# Cookbook:: splunk_otel_collector
# Recipe:: collector_zypper_repo

if node['splunk_otel_collector']['local_artifact_testing_enabled']

  file_name = 'soc.rpm'
  rpm_install_dir = '/etc/otel/collector/'
  rpm_install_path = rpm_install_dir + file_name

  # Create destination dir on target machine
  directory rpm_install_dir do
    mode '0644'
    action :create
    recursive true
  end

  # Copy rpm file from source to target
  cookbook_file rpm_install_path do
    source file_name
    mode '0644'
  end

  zypper_package 'splunk-otel-collector' do
    source rpm_install_path
    gpg_check false
    action :install
  end

else

  zypper_repository 'splunk-otel-collector' do
    description 'Splunk OpenTelemetry Collector Repository'
    baseurl "#{node['splunk_otel_collector']['rhel_repo_url']}/#{node['splunk_otel_collector']['package_stage']}/$basearch/"
    gpgcheck true
    gpgkey node['splunk_otel_collector']['rhel_gpg_key_url']
    type 'yum'
    enabled true
    action :create
    notifies :run, 'execute[add-rhel-key]', :immediately
  end

end

execute 'add-rhel-key' do
  command "rpm --import #{node['splunk_otel_collector']['rhel_gpg_key_url']}"
  action :nothing
end
