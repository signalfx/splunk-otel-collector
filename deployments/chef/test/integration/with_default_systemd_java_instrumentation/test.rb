require_relative '../shared/verify_instrumentation'

verify_instrumentation(sdks: %w(java), systemd: true)
