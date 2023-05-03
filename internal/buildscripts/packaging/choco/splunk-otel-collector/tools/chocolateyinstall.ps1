$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
. $toolsDir\common.ps1

write-host "Checking configuration parameters ..."
$pp = Get-PackageParameters

[bool]$WITH_FLUENTD = $TRUE
[bool]$SkipFluentd = $FALSE

$MODE = $pp['MODE']

$SPLUNK_ACCESS_TOKEN = $pp['SPLUNK_ACCESS_TOKEN']
$SPLUNK_INGEST_URL = $pp['SPLUNK_INGEST_URL']
$SPLUNK_API_URL = $pp['SPLUNK_API_URL']
$SPLUNK_HEC_TOKEN = $pp['SPLUNK_HEC_TOKEN']
$SPLUNK_HEC_URL = $pp['SPLUNK_HEC_URL']
$SPLUNK_TRACE_URL = $pp['SPLUNK_TRACE_URL']
$SPLUNK_BUNDLE_DIR = $pp['SPLUNK_BUNDLE_DIR']
$SPLUNK_REALM = $pp['SPLUNK_REALM']
$SPLUNK_MEMORY_TOTAL_MIB = $pp['SPLUNK_MEMORY_TOTAL_MIB']

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

$regkey = "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment"

# default values if parameters not passed
try{
    if (!$SPLUNK_ACCESS_TOKEN) {
        $SPLUNK_ACCESS_TOKEN = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_ACCESS_TOKEN" -ErrorAction SilentlyContinue
    }
    else{
        update_registry -path "$regkey" -name "SPLUNK_ACCESS_TOKEN" -value "$SPLUNK_ACCESS_TOKEN"
    }
}
catch{
    write-host "The SPLUNK_ACCESS_TOKEN parameter is not specified."
}

try {
    if (!$SPLUNK_REALM) {
        $SPLUNK_REALM = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_REALM" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_REALM = "us0"
    write-host "The SPLUNK_REALM parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_INGEST_URL) {
        $SPLUNK_INGEST_URL =  Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_INGEST_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_INGEST_URL = "https://ingest.$SPLUNK_REALM.signalfx.com"
    write-host "The SPLUNK_INGEST_URL parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_API_URL) {
        $SPLUNK_API_URL = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_API_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_API_URL = "https://api.$SPLUNK_REALM.signalfx.com"
    write-host "The SPLUNK_INGEST_URL parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_HEC_TOKEN) {
        $SPLUNK_HEC_TOKEN = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_HEC_TOKEN" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_HEC_TOKEN = $SPLUNK_ACCESS_TOKEN
    write-host "The SPLUNK_HEC_TOKEN parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_HEC_URL) {
        $SPLUNK_HEC_URL = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_HEC_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_HEC_URL = "https://ingest.$SPLUNK_REALM.signalfx.com/v1/log"
    write-host "The SPLUNK_HEC_URL parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_TRACE_URL) {
        $SPLUNK_TRACE_URL = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_TRACE_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_TRACE_URL = "https://ingest.$SPLUNK_REALM.signalfx.com/v2/trace"
    write-host "The SPLUNK_TRACE_URL parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_MEMORY_TOTAL_MIB) {
        $SPLUNK_MEMORY_TOTAL_MIB = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_MEMORY_TOTAL_MIB" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_MEMORY_TOTAL_MIB = "512"
    write-host "The SPLUNK_MEMORY_TOTAL_MIB parameter is not specified. Using default configuration."
}

try {
    if (!$SPLUNK_BUNDLE_DIR) {
        $SPLUNK_BUNDLE_DIR = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_BUNDLE_DIR" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_BUNDLE_DIR = "$installation_path\agent-bundle"
    write-host "The SPLUNK_BUNDLE_DIR parameter is not specified. Using default configuration."
}

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

update_registry -path "$regkey" -name "SPLUNK_API_URL" -value "$SPLUNK_API_URL"
update_registry -path "$regkey" -name "SPLUNK_BUNDLE_DIR" -value "$SPLUNK_BUNDLE_DIR"
update_registry -path "$regkey" -name "SPLUNK_HEC_TOKEN" -value "$SPLUNK_HEC_TOKEN"
update_registry -path "$regkey" -name "SPLUNK_HEC_URL" -value "$SPLUNK_HEC_URL"
update_registry -path "$regkey" -name "SPLUNK_INGEST_URL" -value "$SPLUNK_INGEST_URL"
update_registry -path "$regkey" -name "SPLUNK_MEMORY_TOTAL_MIB" -value "$SPLUNK_MEMORY_TOTAL_MIB"
update_registry -path "$regkey" -name "SPLUNK_REALM" -value "$SPLUNK_REALM"
update_registry -path "$regkey" -name "SPLUNK_TRACE_URL" -value "$SPLUNK_TRACE_URL"

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

update_registry -path "$regkey" -name "SPLUNK_CONFIG" -value "$config_path"

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

# Try starting the service(s) only after all components were successfully installed and SPLUNK_ACCESS_TOKEN was found.
if (!$SPLUNK_ACCESS_TOKEN) {
    write-host ""
    write-host "*NOTICE*: SPLUNK_ACCESS_TOKEN was not specified as an installation parameter and not found in the Windows Registry."
    write-host "This is required for the default configuration to reach Splunk Observability Cloud and can be configured with:"
    write-host "  PS> Set-ItemProperty -path `"HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`" -name `"SPLUNK_ACCESS_TOKEN`" -value `"ACTUAL_ACCESS_TOKEN`""
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
