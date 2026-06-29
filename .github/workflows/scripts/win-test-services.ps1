param (
    [string]$mode = "agent",
    [string]$access_token = "testing123",
    [string]$realm = "test",
    [string]$memory = "512",
    [string]$with_msi_uninstall_comments = "",
    [string]$api_url = "https://api.${realm}.observability.splunkcloud.com",
    [string]$ingest_url = "https://ingest.${realm}.observability.splunkcloud.com",
    [string]$with_svc_args = "",
    [string]$splunk_platform_url = "",
    [string]$splunk_platform_token = "",
    [string]$splunk_platform_logs_index = ""
)

$ErrorActionPreference = 'Stop'
Set-PSDebug -Trace 1

function check_collector_svc_environment([hashtable]$expected_env_vars) {
    $actual_env_vars = @{}
    $env_array = @()
    try {
        $env_array = Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -Name "Environment"
    } catch {
        Write-Host "Assuming an old version of the collector with environment variables at the machine scope"
        $actual_env_vars = [Environment]::GetEnvironmentVariables("Machine")<#Do this if a terminating exception happens#>
    }

    foreach ($entry in $env_array) {
        $key, $value = $entry.Split("=", 2)
        if ($actual_env_vars.ContainsKey($key)) {
            throw "Environment variable $key is duplicated in the splunk-otel-collector service Environment registry entry."
        }
        $actual_env_vars.Add($key, $value)
    }

    foreach ($key in $expected_env_vars.Keys) {
        $expected_value = $expected_env_vars[$key]
        $actual_value = $actual_env_vars[$key]
        if ($expected_value -ne $actual_value) {
            throw "Environment variable $key is not properly set. Found: '$actual_value', Expected '$expected_value'"
        }
    }

    return $actual_env_vars
}

function service_running([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).State -Eq "Running")
}

function append_svc_arg([string]$svc_args, [string]$arg) {
    if ([string]::IsNullOrWhitespace($svc_args)) {
        return $arg
    }
    return "$svc_args $arg"
}

$program_data_collector_dir = "${env:PROGRAMDATA}\Splunk\OpenTelemetry Collector"
$program_files_collector_dir = "${Env:ProgramFiles}\Splunk\OpenTelemetry Collector"
$default_config_path = "${program_data_collector_dir}\${mode}_config.yaml"
$logs_config_path = "${program_data_collector_dir}\splunk_logs_config_windows.yaml"

$expected_svc_env_vars = @{
  "SPLUNK_ACCESS_TOKEN"     = "$access_token";
  "SPLUNK_REALM"            = "$realm";
  "SPLUNK_API_URL"          = "$api_url";
  "SPLUNK_INGEST_URL"       = "$ingest_url";
  "SPLUNK_HEC_URL"          = "${ingest_url}/v1/log";
  "SPLUNK_HEC_TOKEN"        = "$access_token";
}

if (![string]::IsNullOrWhitespace($memory)) {
    $expected_svc_env_vars["SPLUNK_MEMORY_TOTAL_MIB"] = "$memory"
}

if (![string]::IsNullOrWhitespace($splunk_platform_url)) {
    $expected_svc_env_vars["SPLUNK_PLATFORM_URL"] = "$splunk_platform_url"
}
if (![string]::IsNullOrWhitespace($splunk_platform_token)) {
    $expected_svc_env_vars["SPLUNK_PLATFORM_TOKEN"] = "$splunk_platform_token"
}
if (![string]::IsNullOrWhitespace($splunk_platform_logs_index)) {
    $expected_svc_env_vars["SPLUNK_PLATFORM_LOGS_INDEX"] = "$splunk_platform_logs_index"
}

$actual_svc_env_vars = check_collector_svc_environment $expected_svc_env_vars
$actual_config_path = ""
if ($actual_svc_env_vars.ContainsKey("SPLUNK_CONFIG")) {
    $actual_config_path = $actual_svc_env_vars["SPLUNK_CONFIG"]
}
$config_set_by_env = ![string]::IsNullOrWhitespace($actual_config_path)
if ($config_set_by_env -and ($actual_config_path -ne $default_config_path)) {
    throw "Environment variable SPLUNK_CONFIG is not properly set. Found: '$actual_config_path', Expected '$default_config_path'"
}

if ((service_running -name "splunk-otel-collector")) {
    write-host "splunk-otel-collector service is running."
} else {
    throw "splunk-otel-collector service is not running."
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
    if (Test-Path -Path "${program_files_collector_dir}\*_config.yaml") {
        throw "Found config files in '${program_files_collector_dir}' these files should not be present"
    }
}

If (!(Test-Path -Path "$default_config_path")) {
    throw "Config file '$default_config_path' was not found after the install"
}

if (![string]::IsNullOrWhitespace($splunk_platform_url) -and !(Test-Path -Path "$logs_config_path")) {
    throw "Config file '$logs_config_path' was not found after the install"
}

$svc_commandline = ""
try {
    $svc_commandline = Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -Name "ImagePath"
} catch {
    throw "Failed to retrieve the service command line from the registry."
}

$expected_svc_args = $with_svc_args.Trim('"').Replace('""', '"')
if (!$config_set_by_env -and ![string]::IsNullOrWhitespace($access_token)) {
    $expected_svc_args = append_svc_arg $expected_svc_args "--config `"${default_config_path}`""
}
if (![string]::IsNullOrWhitespace($splunk_platform_url)) {
    $expected_svc_args = append_svc_arg $expected_svc_args "--config `"${logs_config_path}`""
}
if (!$config_set_by_env -and ![string]::IsNullOrWhitespace($access_token) -and ![string]::IsNullOrWhitespace($splunk_platform_url)) {
    $expected_svc_args = append_svc_arg $expected_svc_args "--feature-gates=confmap.enableMergeAppendOption"
}

if ($expected_svc_args -ne "") {
    if (-not $svc_commandline.EndsWith($expected_svc_args)) {
        throw "Service command line does not match the expected arguments. Found: '$svc_commandline', Expected to end with: '$expected_svc_args'"
    } else {
        Write-Host "Service command line matches the expected arguments."
    }
}
