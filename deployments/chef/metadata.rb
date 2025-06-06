name 'splunk_otel_collector'
maintainer 'Splunk, Inc.'
maintainer_email 'signalfx-support@splunk.com'
license 'Apache-2.0'
description 'Install/Configure the Splunk OpenTelemetry Collector'
version '0.17.0'
chef_version '>= 16.0'

supports 'amazon'
supports 'centos'
supports 'debian'
supports 'oracle'
supports 'redhat'
supports 'rockylinux'
supports 'suse'
supports 'ubuntu'
supports 'windows'

gem 'rubyzip', '< 2.0.0'

issues_url 'https://github.com/signalfx/splunk-otel-collector/issues'
source_url 'https://github.com/signalfx/splunk-otel-collector'
