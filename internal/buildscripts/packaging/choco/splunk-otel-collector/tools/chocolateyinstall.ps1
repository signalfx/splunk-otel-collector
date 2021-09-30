$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
. $toolsDir\common.ps1

echo "Checking configuration parameters ..."
$pp = Get-PackageParameters

$install_dir = $pp['install_dir']
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
    write-host "The SPLUNK_ACCESS_TOKEN parameter is not speacified."
}

try {
    if (!$SPLUNK_REALM) {
        $SPLUNK_REALM = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_REALM" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_REALM = "us0"
    write-host "The SPLUNK_REALM parameter is not speacified. Using default configuration."
}

try {
    if (!$SPLUNK_INGEST_URL) {
        $SPLUNK_INGEST_URL =  Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_INGEST_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_INGEST_URL = "https://ingest.$SPLUNK_REALM.signalfx.com"
    write-host "The SPLUNK_INGEST_URL parameter is not speacified. Using default configuration."
}

try {
    if (!$SPLUNK_API_URL) {
        $SPLUNK_API_URL = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_API_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_API_URL = "https://api.$SPLUNK_REALM.signalfx.com"
    write-host "The SPLUNK_INGEST_URL parameter is not speacified. Using default configuration."
}

try {
    if (!$SPLUNK_HEC_TOKEN) {
        $SPLUNK_HEC_TOKEN = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_HEC_TOKEN" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_HEC_TOKEN = $SPLUNK_ACCESS_TOKEN
    write-host "The SPLUNK_HEC_TOKEN parameter is not speacified. Using default configuration."
}

try {
    if (!$SPLUNK_HEC_URL) {
        $SPLUNK_HEC_URL = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_HEC_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_HEC_URL = "https://ingest.$SPLUNK_REALM.signalfx.com/v1/log"
    write-host "The SPLUNK_HEC_URL parameter is not speacified. Using default configuration."
}

try {
    if (!$SPLUNK_TRACE_URL) {
        $SPLUNK_TRACE_URL = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_TRACE_URL" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_TRACE_URL = "https://ingest.$SPLUNK_REALM.signalfx.com/v2/trace"
    write-host "The SPLUNK_TRACE_URL parameter is not speacified. Using default configuration."
}

try {
    if (!$SPLUNK_MEMORY_TOTAL_MIB) {
        $SPLUNK_MEMORY_TOTAL_MIB = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_MEMORY_TOTAL_MIB" -ErrorAction SilentlyContinue
    }
}
catch {
    $SPLUNK_MEMORY_TOTAL_MIB = "512"
    write-host "The SPLUNK_MEMORY_TOTAL_MIB parameter is not speacified. Using default configuration."
}

if (!$SPLUNK_BUNDLE_DIR) {
    $SPLUNK_BUNDLE_DIR = "$installation_path\agent-bundle"
    write-host "The SPLUNK_BUNDLE_DIR parameter is not speacified. Using default configuration."
}

if (!$install_dir) {
    $install_dir = $installation_path
    write-host "Setting installation directory to $install_dir"
}

# remove orphaned service or when upgrading from bundle installation
try {
    stop_service
} catch {
    echo "$_"
}

# remove orphaned registry entries or when upgrading from bundle installation
try {
    remove_otel_registry_entries
} catch {
    echo "$_"
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
    silentArgs     = "/qn /norestart /l*v `"$($env:TEMP)\$($packageName).$($env:chocolateyPackageVersion).MsiInstall.log`" INSTALLDIR=`"$($install_dir)`""
    validExitCodes = @(0)
}

Install-ChocolateyInstallPackage @packageArgs

if ($MODE -eq "agent" -or !$MODE) {
    write-host "Copying agent_config.yaml to $config_path"
    Copy-Item "$installation_path\agent_config.yaml" "$config_path"
    $config_path = "$program_data_path\agent_config.yaml"
}
elseif ($MODE -eq "gateway"){
    write-host "Copying gateway_config.yaml to $config_path"
    Copy-Item "$installation_path\gateway_config.yaml" "$config_path"
    $config_path = "$program_data_path\gateway_config.yaml"
}

update_registry -path "$regkey" -name "SPLUNK_CONFIG" -value "$config_path"

if (!$SPLUNK_ACCESS_TOKEN) {
    echo ""
    echo "*NOTICE*: SPLUNK_ACCESS_TOKEN environment variable needs to specify with a valid value. After specifying it to start the splunk-otel-collector service rebooting the system or run the following command in a PowerShell terminal:"
    echo " Start-Service -Name `"splunk-otel-collector`""
    echo ""
} else {
    echo "Starting splunk-otel-collector service..."
    start_service -config_path "$config_path"
    wait_for_service -timeout 60
    echo "- Started"
}
