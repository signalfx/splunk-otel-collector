param (
    [string]$mode = "agent",
    [string]$access_token = "testing123",
    [string]$realm = "test",
    [string]$memory = "512",
    [string]$with_fluentd = "true",
    [string]$with_msi_uninstall_comments = ""
)

$ErrorActionPreference = 'Stop'
Set-PSDebug -Trace 1

function check_collector_svc_environment([hashtable]$expected_env_vars) {
    $actual_env_vars = @{}
    try {
        $env_array = Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -Name "Environment"
        foreach ($entry in $env_array) {
            $key, $value = $entry.Split("=", 2)
            $actual_env_vars.Add($key, $value)
        }
    } catch {
        Write-Host "Assuming an old version of the collector with environment variables at the machine scope"
        $actual_env_vars = [Environment]::GetEnvironmentVariables("Machine")<#Do this if a terminating exception happens#>
    }

    foreach ($key in $expected_env_vars.Keys) {
        $expected_value = $expected_env_vars[$key]
        $actual_value = $actual_env_vars[$key]
        if ($expected_value -ne $actual_value) {
            throw "Environment variable $key is not properly set. Found: '$actual_value', Expected '$expected_value'"
        }
    }
}

function service_running([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).State -Eq "Running")
}

$api_url = "https://api.${realm}.signalfx.com"
$ingest_url = "https://ingest.${realm}.signalfx.com"

check_collector_svc_environment @{
  "SPLUNK_CONFIG"           = "${env:PROGRAMDATA}\Splunk\OpenTelemetry Collector\${mode}_config.yaml";
  "SPLUNK_ACCESS_TOKEN"     = "$access_token";
  "SPLUNK_REALM"            = "$realm";
  "SPLUNK_API_URL"          = "$api_url";
  "SPLUNK_INGEST_URL"       = "$ingest_url";
  "SPLUNK_TRACE_URL"        = "${ingest_url}:443";
  "SPLUNK_HEC_URL"          = "${ingest_url}/v1/log";
  "SPLUNK_HEC_TOKEN"        = "$access_token";
  "SPLUNK_BUNDLE_DIR"       = "${env:PROGRAMFILES}\Splunk\OpenTelemetry Collector\agent-bundle";
  "SPLUNK_MEMORY_TOTAL_MIB" = "$memory";
}

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

if ($with_msi_uninstall_comments -ne "") {
    $uninstallProperties = Get-ChildItem -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" |
        ForEach-Object { Get-ItemProperty $_.PSPath } |
        Where-Object { $_.DisplayName -eq "Splunk OpenTelemetry Collector" }
    if ($with_msi_uninstall_comments -ne $uninstallProperties.Comments) {
        throw "Uninstall Comments in registry are not properly set. Found: '$uninstallProperties.Comments', Expected '$with_msi_uninstall_comments'"
    } else {
        write-host "Uninstall Comments in registry are properly set."
    }
}
