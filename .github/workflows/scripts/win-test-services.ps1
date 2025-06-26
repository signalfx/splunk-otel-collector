param (
    [string]$mode = "agent",
    [string]$access_token = "testing123",
    [string]$realm = "test",
    [string]$memory = "512",
    [string]$with_fluentd = "true",
    [string]$with_msi_uninstall_comments = "",
    [string]$api_url = "https://api.${realm}.signalfx.com",
    [string]$ingest_url = "https://ingest.${realm}.signalfx.com",
    [string]$with_svc_args = ""
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

$expected_svc_env_vars = @{
  "SPLUNK_CONFIG"           = "${env:PROGRAMDATA}\Splunk\OpenTelemetry Collector\${mode}_config.yaml";
  "SPLUNK_ACCESS_TOKEN"     = "$access_token";
  "SPLUNK_REALM"            = "$realm";
  "SPLUNK_API_URL"          = "$api_url";
  "SPLUNK_INGEST_URL"       = "$ingest_url";
  "SPLUNK_HEC_URL"          = "${ingest_url}/v1/log";
  "SPLUNK_HEC_TOKEN"        = "$access_token";
  "SPLUNK_BUNDLE_DIR"       = "${env:PROGRAMFILES}\Splunk\OpenTelemetry Collector\agent-bundle";
}

if (![string]::IsNullOrWhitespace($memory)) {
    $expected_svc_env_vars["SPLUNK_MEMORY_TOTAL_MIB"] = "$memory"
}

check_collector_svc_environment $expected_svc_env_vars

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

$uninstallProperties = Get-ChildItem -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" |
    ForEach-Object { Get-ItemProperty $_.PSPath } |
    Where-Object { $_.DisplayName -eq "Splunk OpenTelemetry Collector" }
if ($with_msi_uninstall_comments -ne "") {
    if ($with_msi_uninstall_comments -ne $uninstallProperties.Comments) {
        throw "Uninstall Comments in registry are not properly set. Found: '$uninstallProperties.Comments', Expected '$with_msi_uninstall_comments'"
    } else {
        write-host "Uninstall Comments in registry are properly set."
    }
}

$installed_version = [Version]$uninstallProperties.DisplayVersion
if ($installed_version -gt [Version]"0.97.0.0") {
    if (Test-Path -Path "${Env:ProgramFiles}\Splunk\OpenTelemetry Collector\*_config.yaml") {
        throw "Found config files in '${Env:ProgramFiles}\Splunk\OpenTelemetry Collector' these files should not be present"
    }
}

If (!(Test-Path -Path "${Env:ProgramData}\Splunk\OpenTelemetry Collector\*_config.yaml")) {
    throw "No config files found in ${Env:ProgramData}\Splunk\OpenTelemetry Collector these files are expected after the install"
}

$svc_commandline = ""
try {
    $svc_commandline = Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -Name "ImagePath"
} catch {
    throw "Failed to retrieve the service command line from the registry."
}

if ($with_svc_args -ne "") {
    if ($svc_commandline.EndsWith($with_svc_args)) {
        throw "Service command line does not match the expected arguments. Found: '$svc_commandline', Expected to end with: '$with_svc_args'"
    } else {
        Write-Host "Service command line matches the expected arguments."
    }
}
