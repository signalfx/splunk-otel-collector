$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
. $toolsDir\common.ps1

write-host "Checking configuration parameters ..."
$pp = Get-PackageParameters

[bool]$WITH_FLUENTD = $FALSE
[bool]$SkipFluentd = $FALSE

$MODE = $pp['MODE']

if ($MODE -and ($MODE -ne "agent") -and ($MODE -ne "gateway")) {
    throw "Invalid value of MODE option is specified. Collector service can only run in agent or gateway mode."
}

if ($pp['WITH_FLUENTD']) {
    if (($pp['WITH_FLUENTD'] -eq "true") -or ($pp['WITH_FLUENTD'] -eq "false")){
        try {
            $WITH_FLUENTD = [System.Convert]::ToBoolean($pp['WITH_FLUENTD']) 
        } catch [FormatException] {
            $WITH_FLUENTD = $FALSE
        }
    }
    else {
        throw "Invalid value of WITH_FLUENTD option is specified. Possible values are true and false."
    }
}

if ($WITH_FLUENTD) {
    # check execution policy
    Write-Host 'Checking execution policy'
    check_policy
}

# Read splunk-otel-collector environment variables.
$env_vars = @{}
$env_var_names = @(
    "SPLUNK_ACCESS_TOKEN",
    "SPLUNK_REALM",
    "SPLUNK_INGEST_URL",
    "SPLUNK_API_URL",
    "SPLUNK_HEC_TOKEN",
    "SPLUNK_HEC_URL",
    "SPLUNK_TRACE_URL",
    "SPLUNK_MEMORY_TOTAL_MIB",
    "SPLUNK_BUNDLE_DIR",
    "SPLUNK_LISTEN_INTERFACE"
)
$upgraded_from_version_with_machine_wide_env_vars = $false

Write-Host "Checking for previous installation..."
# First check if any previous version of the collector is installed.
$installed_collector = Get-ItemProperty "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*" -ErrorAction SilentlyContinue |
    Where-Object { $_.DisplayName -eq "Splunk OpenTelemetry Collector" }
if ($installed_collector) {
    # The package is already present, so this is an upgrade.
    Write-Host "Found a previous installation..."
    $installed_version = [Version]$installed_collector.DisplayVersion # Version for chocolatey doesn't include the prefilx 'v', this conversion is fine.
    $last_version_with_machine_env_vars = [Version]"0.92.0.0"
    if ($installed_version -le $last_version_with_machine_env_vars) {
        $upgraded_from_version_with_machine_wide_env_vars = $true
        Write-Host "Getting machine wide environment variables..."
        foreach ($name in $env_var_names) {
            $value = [Environment]::GetEnvironmentVariable($name, "Machine")
            if ($value) {
                $env_vars[$name] = "$value"
            }
        }
    }
}

$reg_path = Join-Path "HKLM:\SYSTEM\CurrentControlSet\Services" $service_name
if (Test-Path $reg_path) {
    Write-Host "Service registry entry key found: $reg_path"
    $previous_environment = Get-ItemProperty $reg_path -Name "Environment" -ErrorAction SilentlyContinue
    if ($previous_environment) {
        Write-Host "Found previous environment variables for the $service_name service."
        foreach ($entry in $previous_environment) {
            $k, $v = $entry.Split("=", 2)
            $env_vars[$k] = $v
        }
    }
}

# Use default values if package parameters not set
Write-Host "Setting default values for missing parameters..."
$access_token = $pp["SPLUNK_ACCESS_TOKEN"]
if ($access_token) {
    $env_vars["SPLUNK_ACCESS_TOKEN"] = "$access_token" # Env. var values are always strings
} elseif (!$env_vars["SPLUNK_ACCESS_TOKEN"]) {
    write-host "The SPLUNK_ACCESS_TOKEN parameter is not specified."
}

set_env_var_value_from_package_params $env_vars $pp "SPLUNK_REALM" "us0"
$realm = $env_vars["SPLUNK_REALM"] # Cache the realm since it is used to build various default values.

set_env_var_value_from_package_params $env_vars $pp "SPLUNK_INGEST_URL"         "https://ingest.$realm.signalfx.com"
set_env_var_value_from_package_params $env_vars $pp "SPLUNK_API_URL"            "https://api.$realm.signalfx.com"
set_env_var_value_from_package_params $env_vars $pp "SPLUNK_HEC_TOKEN"          $env_vars["SPLUNK_ACCESS_TOKEN"]
set_env_var_value_from_package_params $env_vars $pp "SPLUNK_HEC_URL"            "https://ingest.$realm.signalfx.com/v1/log"
set_env_var_value_from_package_params $env_vars $pp "SPLUNK_TRACE_URL"          "https://ingest.$realm.signalfx.com/v2/trace"
set_env_var_value_from_package_params $env_vars $pp "SPLUNK_MEMORY_TOTAL_MIB"   "512"
set_env_var_value_from_package_params $env_vars $pp "SPLUNK_BUNDLE_DIR"         "$installation_path\agent-bundle"

# remove orphaned service or when upgrading from bundle installation
if (service_installed -name "$service_name") {
    try {
        stop_service -name "$service_name"
    } catch {
        write-host "$_"
    }
}

# remove orphaned registry entries or when upgrading from bundle installation
try {
    remove_otel_registry_entries
} catch {
    write-host "$_"
}

if ($upgraded_from_version_with_machine_wide_env_vars) {
    # Remove the machine-wide environment variables that were set by previous versions of the collector.
    foreach ($name in $env_var_names) {
        [Environment]::SetEnvironmentVariable($name, $null, "Machine")
    }
}

$packageArgs = @{
    packageName    = $env:ChocolateyPackageName
    fileType       = 'msi'
    file           = Join-Path "$toolsDir" "MSI_NAME"  # replaced at build time
    softwareName   = $env:ChocolateyPackageTitle
    checksum64     = "MSI_HASH"  # replaced at build time
    checksumType64 = 'sha256'
    silentArgs     = "/qn /norestart /l*v `"$($env:TEMP)\$($packageName).$($env:chocolateyPackageVersion).MsiInstall.log`" INSTALLDIR=`"$($installation_path)`""
    validExitCodes = @(0)
}

Install-ChocolateyInstallPackage @packageArgs

if ($MODE -eq "agent" -or !$MODE) {
    $config_path = "$program_data_path\agent_config.yaml"
    if (-NOT (Test-Path -Path "$config_path")) {
        write-host "Copying agent_config.yaml to $config_path"
        Copy-Item "$installation_path\agent_config.yaml" "$config_path"
    }
}
elseif ($MODE -eq "gateway"){
    $config_path = "$program_data_path\gateway_config.yaml"
    if (-NOT (Test-Path -Path "$config_path")) {
        write-host "Copying gateway_config.yaml to $config_path"
        Copy-Item "$installation_path\gateway_config.yaml" "$config_path"
    }
}

$env_vars["SPLUNK_CONFIG"] = "$config_path"

# Install and configure fluentd to forward log events to the collector.
if ($WITH_FLUENTD) {
    # Skip installation of fluentd if already installed
    if ((service_installed -name "$fluentd_service_name") -OR (Test-Path -Path "$fluentd_base_dir\bin\fluentd")) {
        $SkipFluentd = $TRUE
        Write-Host "The $fluentd_service_name service is already installed. Skipping fluentd installation."
    } else {
        . $toolsDir\fluentd.ps1
    }
}

set_service_environment $service_name $env_vars

# Try starting the service(s) only after all components were successfully installed and SPLUNK_ACCESS_TOKEN was found.
if (!$env_vars["SPLUNK_ACCESS_TOKEN"]) {
    write-host ""
    write-host "*NOTICE*: SPLUNK_ACCESS_TOKEN was not specified as an installation parameter and not found in the Windows Registry."
    write-host "This is required for the default configuration to reach Splunk Observability Cloud and can be configured with:"
    write-host '  PS> $values = Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -Name "Environment" | Select-Object -ExpandProperty Environment'
    write-host '  PS> $values += "SPLUNK_ACCESS_TOKEN=<your_access_token>"'
    write-host '  PS> Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -Name $propertyName -Value $values -Type MultiString'
    write-host "before starting the $service_name service with:"
    write-host "  PS> Start-Service -Name `"${service_name}`""
    if ($WITH_FLUENTD) {
        write-host "Then restart the fluentd service to ensure collected log events are forwarded to the $service_name service with:"
        write-host "  PS> Stop-Service -Name `"${fluentd_service_name}`""
        write-host "  PS> Start-Service -Name `"${fluentd_service_name}`""
    }
    write-host ""
} else {
    try {
        write-host "Starting $service_name service..."
        start_service -name "$service_name" -config_path "$config_path"
        write-host "- Started"
        if ($WITH_FLUENTD -and !$SkipFluentd) {
            # The fluentd service is automatically started after msi installation.
            # Wait for it to be running before trying to restart it with our custom config.
            Write-Host "Restarting $fluentd_service_name service..."
            wait_for_service -name "$fluentd_service_name"
            stop_service -name "$fluentd_service_name"
            start_service -name "$fluentd_service_name" -config_path "$fluentd_config_path"
            Write-Host "- Started"
        }
        write-host ""
    } catch {
        $err = $_.Exception.Message
        # Don't fail if all components were installed successfully but service(s) fail to start.
        # Otherwise, chocolatey may leave the system in a weird state.
        Write-Warning "Installation completed, but one or more services failed to start:"
        Write-Warning "$err"
        continue
    }
}
