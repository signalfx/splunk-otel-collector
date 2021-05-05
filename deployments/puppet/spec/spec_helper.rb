require 'rspec-puppet'
require 'rspec-puppet-facts'
require 'puppetlabs_spec_helper/module_spec_helper'

include RspecPuppetFacts

fixture_path = File.join(File.dirname(File.expand_path(__FILE__)), 'fixtures')

RSpec.configure do |c|
  c.module_path     = File.join(fixture_path, 'modules')
  c.manifest_dir    = File.join(fixture_path, 'manifests')
  c.manifest        = File.join(fixture_path, 'manifests', 'site.pp')
  c.environmentpath = File.join(Dir.pwd, 'spec')
  c.default_facts   = { :osfamily => 'redhat', :service_provider => 'systemd' }
end
