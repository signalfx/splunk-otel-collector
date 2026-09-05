require_relative '../shared/verify_instrumentation'

verify_instrumentation(sdks: %w(nodejs), systemd: false, custom: true, custom_preload: true)
