# Cookbook:: splunk_otel_collector
# Recipe:: collector_zypper_repo

zypper_repository 'splunk-otel-collector' do
  description 'Splunk OpenTelemetry Collector Repository'
  baseurl node['splunk_otel_collector']['rhel_repo_url']
  gpgcheck true
  gpgkey node['splunk_otel_collector']['rhel_gpg_key_url']
  type 'yum'
  enabled true
  action :create
  notifies :run, 'execute[add-rhel-key]', :immediately
end

if node['splunk_otel_collector']['local_artifact_testing_enabled']
  file_name = 'soc.rpm'
  rpm_install_path = '/tmp/' + file_name

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
end

execute 'add-rhel-key' do
  command "rpm --import #{node['splunk_otel_collector']['rhel_gpg_key_url']}"
  action :nothing
end
