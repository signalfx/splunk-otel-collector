$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
. $toolsDir\common.ps1


$installation_path = "$drive" + "\Program Files\Splunk\OpenTelemetry Collector"
$program_data_path = "$drive" + "\ProgramData\Splunk\OpenTelemetry Collector"
$config_path = "$program_data_path\agent_config.yaml"

echo "Checking configuration parameters ..."
$pp = Get-PackageParameters

$SPLUNK_ACCESS_TOKEN = $pp['SPLUNK_ACCESS_TOKEN']
$SPLUNK_INGEST_URL = $pp['SPLUNK_INGEST_URL']
$SPLUNK_API_URL = $pp['SPLUNK_API_URL']
$install_dir = $pp['install_dir']
$SPLUNK_HEC_TOKEN = $pp['SPLUNK_HEC_TOKEN']
$SPLUNK_HEC_URL = $pp['SPLUNK_HEC_URL']
$SPLUNK_TRACE_URL = $pp['SPLUNK_TRACE_URL']
$SPLUNK_BUNDLE_DIR = $pp['SPLUNK_BUNDLE_DIR']
$SPLUNK_REALM = $pp['SPLUNK_REALM']


# create config files
create_program_data

# get param values from config files if they exist
if (!$SPLUNK_ACCESS_TOKEN) {
    $SPLUNK_ACCESS_TOKEN = get_value_from_file -path "$program_data_path\token"
    if (!$SPLUNK_ACCESS_TOKEN) {
        echo "The 'SPLUNK_ACCESS_TOKEN' parameter was not specified."
        $SPLUNK_ACCESS_TOKEN = ""
    } else {
        echo "Using access token from $program_data_path\token"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\token", "$SPLUNK_ACCESS_TOKEN", [System.Text.Encoding]::ASCII)

# get param values from config files if they exist
if (!$SPLUNK_REALM) {
    $SPLUNK_REALM = get_value_from_file -path "$program_data_path\realm"
    if (!$SPLUNK_REALM) {
        echo "The 'SPLUNK_REALM' parameter was not specified."
        $SPLUNK_REALM = "us0"
    } else {
        echo "Using a realm from $program_data_path\realm"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\realm", "$SPLUNK_REALM", [System.Text.Encoding]::ASCII)


if (!$SPLUNK_INGEST_URL) {
    $SPLUNK_INGEST_URL = get_value_from_file -path "$program_data_path\SPLUNK_INGEST_URL"
    if (!$SPLUNK_INGEST_URL) {
        $SPLUNK_INGEST_URL = "https://ingest.$SPLUNK_REALM.signalfx.com"
        echo "Setting ingest url to $SPLUNK_INGEST_URL"
    } else {
        echo "Using ingest url from $program_data_path\SPLUNK_INGEST_URL"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\SPLUNK_INGEST_URL", "$SPLUNK_INGEST_URL", [System.Text.Encoding]::ASCII)

if (!$SPLUNK_API_URL) {
    $SPLUNK_API_URL = get_value_from_file -path "$program_data_path\SPLUNK_API_URL"
    if (!$SPLUNK_API_URL) {
        $SPLUNK_API_URL = "https://api.$SPLUNK_REALM.signalfx.com"
        echo "Setting api url to $SPLUNK_API_URL"
    } else {
        echo "Using api url from $program_data_path\SPLUNK_API_URL"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\SPLUNK_API_URL", "$SPLUNK_API_URL", [System.Text.Encoding]::ASCII)

if (!$install_dir) {
    $install_dir = $installation_path
}
echo "Setting installation directory to $install_dir"

if (!$SPLUNK_HEC_TOKEN) {
    $SPLUNK_HEC_TOKEN = get_value_from_file -path "$program_data_path\hec-token"
    if (!$SPLUNK_HEC_TOKEN) {
        echo "The 'SPLUNK_HEC_TOKEN' parameter was not specified."
        $SPLUNK_HEC_TOKEN = ""
    } else {
        echo "Using hec token from $program_data_path\hec-token"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\hec-token", "$SPLUNK_HEC_TOKEN", [System.Text.Encoding]::ASCII)

if (!$SPLUNK_HEC_URL) {
    $SPLUNK_HEC_URL = get_value_from_file -path "$program_data_path\SPLUNK_HEC_URL"
    if (!$SPLUNK_HEC_URL) {
        $SPLUNK_HEC_URL = "https://ingest.$SPLUNK_REALM.signalfx.com/v1/log"
        echo "Setting hec url to $SPLUNK_HEC_URL"
    } else {
        echo "Using hec url from $program_data_path\SPLUNK_HEC_URL"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\SPLUNK_HEC_URL", "$SPLUNK_HEC_URL", [System.Text.Encoding]::ASCII)

if (!$SPLUNK_BUNDLE_DIR) {
    $SPLUNK_BUNDLE_DIR = get_value_from_file -path "$program_data_path\SPLUNK_BUNDLE_DIR"
    if (!$SPLUNK_BUNDLE_DIR) {
        $SPLUNK_BUNDLE_DIR = "$installation_path\agent-bundle"
        echo "Setting bundle dir to $SPLUNK_BUNDLE_DIR"
    } else {
        echo "Using bundle dir from $program_data_path\SPLUNK_BUNDLE_DIR"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\SPLUNK_HEC_URL", "$SPLUNK_HEC_URL", [System.Text.Encoding]::ASCII)

if (!$SPLUNK_TRACE_URL) {
    $SPLUNK_TRACE_URL = get_value_from_file -path "$program_data_path\SPLUNK_TRACE_URL"
    if (!$SPLUNK_TRACE_URL) {
        $SPLUNK_TRACE_URL = "https://ingest.$SPLUNK_REALM.signalfx.com/v2/trace"
        echo "Setting trace url to $SPLUNK_TRACE_URL"
    } else {
        echo "Using trace url from $program_data_path\SPLUNK_TRACE_URL"
    }
}
[System.IO.File]::WriteAllText("$program_data_path\SPLUNK_TRACE_URL", "$SPLUNK_TRACE_URL", [System.Text.Encoding]::ASCII)

$regkey = "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment"
update_registry -path "$regkey" -name "SPLUNK_ACCESS_TOKEN" -value "$SPLUNK_ACCESS_TOKEN"
update_registry -path "$regkey" -name "SPLUNK_API_URL" -value "$SPLUNK_API_URL"
update_registry -path "$regkey" -name "SPLUNK_BUNDLE_DIR" -value "$SPLUNK_BUNDLE_DIR"
update_registry -path "$regkey" -name "SPLUNK_CONFIG" -value "$config_path"
update_registry -path "$regkey" -name "SPLUNK_HEC_TOKEN" -value "$SPLUNK_HEC_TOKEN"
update_registry -path "$regkey" -name "SPLUNK_HEC_URL" -value "$SPLUNK_HEC_URL"
update_registry -path "$regkey" -name "SPLUNK_INGEST_URL" -value "$SPLUNK_INGEST_URL"
update_registry -path "$regkey" -name "SPLUNK_MEMORY_TOTAL_MIB" -value ""
update_registry -path "$regkey" -name "SPLUNK_REALM" -value "$SPLUNK_REALM"
update_registry -path "$regkey" -name "SPLUNK_TRACE_URL" -value "$SPLUNK_TRACE_URL"

$packageArgs = @{
    packageName    = $env:ChocolateyPackageName
    fileType       = 'msi'
    file           = Get-Item "$toolsDir\*amd64.msi"  # replaced at build time
    softwareName   = $env:ChocolateyPackageTitle
    checksum64     = "MSI_HASH"  # replaced at build time
    checksumType64 = 'sha256'
    silentArgs     = "/qn /norestart /l*v `"$($env:TEMP)\$($packageName).$($env:chocolateyPackageVersion).MsiInstall.log`" INSTALLDIR=`"$($install_dir)`""
    validExitCodes = @(0)
}

Install-ChocolateyInstallPackage @packageArgs

if (!(Test-Path -Path "$config_path")) {
    echo "$config_path not found"
    echo "Copying default agent_config.yaml to $config_path"
    Copy-Item "$install_dir\Splunk\OpenTelemetry Collector\agent_config.yaml" "$config_path"
}




