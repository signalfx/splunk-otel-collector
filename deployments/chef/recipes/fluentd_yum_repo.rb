#
# Cookbook:: splunk-otel-collector
# Recipe:: collector_yum_repo
#
# Copyright:: 2021, The Authors, All Rights Reserved.

td_agent_major_version = node['splunk-otel-collector']['fluentd_version'].split('.')[0]

distro = if platform_family?('amazon')
           'amazon'
         else
           'redhat'
         end

package %w(libcap-ng libcap-ng-devel pkgconfig)

execute 'devtools' do
  command 'yum -y groupinstall "Development Tools"'
end

execute 'add-rhel-key' do
  command "rpm --import #{node['splunk-otel-collector']['fluentd_base_url']}/GPG-KEY-td-agent"
end

yum_repository 'splunk-otel-collector' do
  description 'TreasureData Repository'
  baseurl "#{node['splunk-otel-collector']['fluentd_base_url']}/#{td_agent_major_version}/#{distro}/$releasever/$basearch"
  gpgcheck true
  gpgkey "#{node['splunk-otel-collector']['fluentd_base_url']}/GPG-KEY-td-agent"
  enabled true
  action :create
end
