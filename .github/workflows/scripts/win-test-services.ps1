param (
    [string]$mode = "agent",
    [string]$access_token = "testing123",
    [string]$realm = "test",
    [string]$memory = "512",
    [string]$with_fluentd = "true"
)

$ErrorActionPreference = 'Stop'
Set-PSDebug -Trace 1

function check_regkey([string]$name, [string]$value) {
    $actual = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "$name"
    if ( "$value" -ne "$actual" ) {
        throw "Environment variable $name is not properly set. Found: '$actual', Expected '$value'"
    }
}

function service_running([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).State -Eq "Running")
}

$api_url = "https://api.${realm}.signalfx.com"
$ingest_url = "https://ingest.${realm}.signalfx.com"

check_regkey -name "SPLUNK_CONFIG" -value "${env:PROGRAMDATA}\Splunk\OpenTelemetry Collector\${mode}_config.yaml"
check_regkey -name "SPLUNK_ACCESS_TOKEN" -value "$access_token"
check_regkey -name "SPLUNK_REALM" -value "$realm"
check_regkey -name "SPLUNK_API_URL" -value "$api_url"
check_regkey -name "SPLUNK_INGEST_URL" -value "$ingest_url"
check_regkey -name "SPLUNK_TRACE_URL" -value "${ingest_url}/v2/trace"
check_regkey -name "SPLUNK_HEC_URL" -value "${ingest_url}/v1/log"
check_regkey -name "SPLUNK_HEC_TOKEN" -value "$access_token"
check_regkey -name "SPLUNK_BUNDLE_DIR" -value "${env:PROGRAMFILES}\Splunk\OpenTelemetry Collector\agent-bundle"
check_regkey -name "SPLUNK_MEMORY_TOTAL_MIB" -value  "$memory"

if ((service_running -name "splunk-otel-collector")) {
    write-host "splunk-otel-collector service is running."
} else {
    throw "splunk-otel-collector service is not running."
}

if ("$with_fluentd" -eq "true") {
    if ((service_running -name "fluentdwinsvc")) {
        write-host "fluentdwinsvc service is running."
    } else {
        throw "fluentdwinsvc service is not running."
    }
} else {
    if ((service_running -name "fluentdwinsvc")) {
        throw "fluentdwinsvc service is running."
    } else {
        write-host "fluentdwinsvc service is not running."
    }
}
