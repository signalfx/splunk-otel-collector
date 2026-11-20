require_relative '../shared/verify_instrumentation'

verify_instrumentation(sdks: %w(nodejs), systemd: true)
