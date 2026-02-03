module github.com/splunk/splunk_otel_dotnet_deployer

go 1.25.5

require github.com/signalfx/splunk-otel-collector v0.0.0

replace github.com/signalfx/splunk-otel-collector => ../..
