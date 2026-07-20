# frozen_string_literal: true

nodejs_install 'nodejs' do
  install_method 'binary'
  version '18.20.8'
  binary_checksums(
    'linux_x64' => 'c9193e6c414891694759febe846f4f023bf48410a6924a8b1520c46565859665'
  )
  append_env_path false
end
