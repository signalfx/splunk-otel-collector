require_relative '../shared/verify_instrumentation'

verify_instrumentation(sdks: %w(java nodejs dotnet), systemd: false, custom: true, custom_preload: true)
