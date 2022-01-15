#
# Cookbook:: splunk-otel-collector
# Recipe:: collector_yum_repo
#
# Copyright:: 2021, The Authors, All Rights Reserved.

if platform_family?('suse')
  is_suse = true
  repo_path = '/etc/zypp/repos.d'
else
  is_suse = false
  repo_path = '/etc/yum.repos.d'
end

if !is_suse
  yum_repository 'splunk-otel-collector' do
    description 'Splunk OpenTelemetry Collector Repository'
    baseurl "#{node['splunk-otel-collector']['rhel_repo_url']}/#{node['splunk-otel-collector']['package_stage']}/$basearch/"
    gpgcheck true
    gpgkey node['splunk-otel-collector']['rhel_gpg_key_url']
    enabled true
    action :create
  end

  package [ 'libcap' ]

else
  file "#{repo_path}/splunk-otel-collector.repo" do
    content <<-EOH
[splunk-otel-collector]
name=Splunk OpenTelemetry Collector Repository
baseurl=#{node['splunk-otel-collector']['rhel_repo_url']}/#{node['splunk-otel-collector']['package_stage']}/$basearch/
gpgcheck=1
gpgkey=#{node['splunk-otel-collector']['rhel_gpg_key_url']}
enabled=1
    EOH
    mode '0644'
    notifies :run, 'execute[add-rhel-key]', :immediately
  end

  execute 'add-rhel-key' do
    command "rpm --import #{node['splunk-otel-collector']['rhel_gpg_key_url']}"
    action :nothing
  end

  if is_suse
    package [ 'libcap-progs' ]

    execute 'zypper-clean' do
      command 'zypper -n clean -a -r splunk-otel-collector'
    end

    execute 'zypper-refresh' do
      command 'zypper -n refresh -r splunk-otel-collector'
    end
  else
    package [ 'libcap' ]

    execute 'yum-clean' do
      command "yum clean all --disablerepo='*' --enablerepo='splunk-otel-collector'"
    end

    execute 'yum-metadata-refresh' do
      command "yum -q -y makecache --disablerepo=* --enablerepo='splunk-otel-collector'"
    end
  end
end
