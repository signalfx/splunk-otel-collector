# frozen_string_literal: true

nodejs_install 'nodejs' do
  install_method 'binary'
  version '18.20.8'
  binary_checksums(
    'linux_x64' => '27a9f3f14d5e99ad05a07ed3524ba3ee92f8ff8b6db5ff80b00f9feb5ec8097a'
  )
  append_env_path false
end
