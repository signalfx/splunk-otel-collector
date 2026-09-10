require_relative '../shared/verify_instrumentation'

verify_instrumentation(sdks: %w(java dotnet), systemd: true)
