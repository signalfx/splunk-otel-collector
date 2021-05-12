# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The following comment block acts as usage for powershell scripts
# you can view it by passing the script as an argument to the cmdlet 'Get-Help'
# To view the paremeter documentation invoke Get-Help with the option '-Detailed'
# ex. PS C:\> Get-Help "<path to script>\install.ps1" -Detailed

<#
.SYNOPSIS
    Installs the Splunk OpenTelemetry Collector from the package repos.
.DESCRIPTION
    Installs the Splunk OpenTelemetry Collector from the package repos. If access_token is not
    provided, it will be prompted for on the console. If you want to view full documentation
    execute Get-Help with the parameter "-Full".
.PARAMETER access_token
    The token used to send metric data to Splunk.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN"
.PARAMETER realm
    (OPTIONAL) The Splunk realm to use (default: "us0"). The ingest, API, trace, and HEC endpoint URLs will automatically be inferred by this value.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -realm "us1"
.PARAMETER memory
    (OPTIONAL) Total memory in MIB to allocate to the collector; automatically calculates the ballast size (default: "512").
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -memory 1024
.PARAMETER mode
    (OPTIONAL) Configure the collector service to run in "agent" or "gateway" mode (default: "agent").
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -mode "gateway"
.PARAMETER ingest_url
    (OPTIONAL) Set the base ingest URL explicitly instead of the URL inferred from the specified realm (default: https://ingest.REALM.signalfx.com).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -ingest_url "https://ingest.us1.signalfx.com"
.PARAMETER api_url
    (OPTIONAL) Set the base API URL explicitly instead of the URL inferred from the specified realm (default: https://api.REALM.signalfx.com).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -api_url "https://api.us1.signalfx.com"
.PARAMETER trace_url
    (OPTIONAL) Set the trace endpoint URL explicitly instead of the endpoint inferred from the specified realm (default: https://ingest.REALM.signalfx.com/v2/trace).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -trace_url "https://ingest.us1.signalfx.com/v2/trace"
.PARAMETER hec_url
    (OPTIONAL) Set the HEC endpoint URL explicitly instead of the endpoint inferred from the specified realm (default: https://ingest.REALM.signalfx.com/v1/log).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -hec_url "https://ingest.us1.signalfx.com/v1/log"
.PARAMETER hec_token
    (OPTIONAL) Set the HEC token if different than the specified Splunk access_token.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -hec_token "HECTOKEN"
.PARAMETER with_fluentd
    (OPTIONAL) Whether to install and configure fluentd to forward log events to the collector (default: $true)
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -with_fluentd $false
.PARAMETER bundle_dir
    (OPTIONAL) The location of your Smart Agent bundle for monitor functionality (default: C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle)
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -bundle_dir "C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle"
.PARAMETER insecure
    (OPTIONAL) If true then certificates will not be checked when downloading resources. Defaults to '$false'.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -insecure $true
.PARAMETER collector_version
    (OPTIONAL) Specify a specific version of the collector to install.  Defaults to the latest version available.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -collector_version "1.2.3"
.PARAMETER stage
    (OPTIONAL) The package stage to install from ['test', 'beta', 'release']. Defaults to 'release'.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -stage "test"
.PARAMETER msi_path
    (OPTIONAL) Specify a local path to a Splunk OpenTelemetry Collector MSI package to install instead of downloading the package.
    If specified, the -collector_version and -stage parameters will be ignored.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -msi_path "C:\SOME_FOLDER\splunk-otel-collector-1.2.3-amd64.msi"
#>

param (
    [parameter(Mandatory=$true)][string]$access_token = "",
    [string]$realm = "us0",
    [string]$memory = "512",
    [ValidateSet('agent','gateway')][string]$mode = "agent",
    [string]$ingest_url = "",
    [string]$api_url = "",
    [string]$trace_url = "",
    [string]$hec_url = "",
    [string]$hec_token = "",
    [bool]$insecure = $false,
    [string]$collector_version = "",
    [bool]$with_fluentd = $true,
    [string]$bundle_dir = "",
    [ValidateSet('test','beta','release')][string]$stage = "release",
    [string]$msi_path = "",
    [bool]$UNIT_TEST = $false
)

$arch = "amd64"
$format = "msi"
$service_name = "splunk-otel-collector"
$signalfx_dl = "https://dl.signalfx.com"
$installation_path = "\Program Files"
$program_data_path = "\ProgramData\Splunk\OpenTelemetry Collector"
$old_config_path = "$program_data_path\config.yaml"
$agent_config_path = "$program_data_path\agent_config.yaml"
$gateway_config_path = "$program_data_path\gateway_config.yaml"
$config_path = ""
$tempdir = "\tmp\Splunk\OpenTelemetry Collector"
$regkey = "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment"

$fluentd_msi_name = "td-agent-4.0.1-x64.msi"
$fluentd_dl_url = "https://packages.treasuredata.com/4/windows/$fluentd_msi_name"
$fluentd_base_dir = "\opt\td-agent"
$fluentd_config_dir = "$fluentd_base_dir\etc\td-agent"
$fluentd_config_path = "$fluentd_config_dir\td-agent.conf"
$fluentd_service_name = "fluentdwinsvc"

# check that we're not running with a restricted execution policy
function check_policy() {
    $executionPolicy  = (Get-ExecutionPolicy)
    $executionRestricted = ($executionPolicy -eq "Restricted")
    if ($executionRestricted) {
        throw @"
Your execution policy is $executionPolicy, this means you will not be able import or use any scripts including modules.
To fix this change you execution policy to something like RemoteSigned.
        PS> Set-ExecutionPolicy RemoteSigned
For more information execute:
        PS> Get-Help about_execution_policies
"@
    }
}

# check if running as administrator
function check_if_admin() {
    $identity = [Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()
    if (-NOT $identity.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        return $false
    }
    return $true
}

# get latest package tag given a stage and format
function get_latest([string]$stage=$stage,[string]$format=$format) {
    $latest_url = "$signalfx_dl/splunk-otel-collector/$format/$stage/latest.txt"
    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        $latest = (New-Object System.Net.WebClient).DownloadString($latest_url).Trim()
    } catch {
        $err = $_.Exception.Message
        $message = "
        An error occurred while fetching the latest package version $latest_url
        $err
        "
        throw "$message"
    }
    return $latest
}

# builds the filename for the package
function get_filename([string]$tag="",[string]$format=$format,[string]$arch=$arch) {
    $filename = "splunk-otel-collector-$tag-$arch.$format"
    return $filename
}

# builds the url for the package
function get_url([string]$stage="", [string]$format=$format, [string]$filename="") {
    return "$signalfx_dl/splunk-otel-collector/$format/$stage/$filename"
}

# download a file to a given destination
function download_file([string]$url, [string]$outputDir, [string]$fileName) {
    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        (New-Object System.Net.WebClient).DownloadFile($url, "$outputDir\$fileName")
    } catch {
        $err = $_.Exception.Message
        $message = "
        An error occurred while downloading $url
        $err
        "
        throw "$message"
    }
}

# ensure a file exists and raise an exception if it doesn't
function ensure_file_exists([string]$path="C:\") {
    if (!(Test-Path -Path "$path")){
        throw "Cannot find the path '$path'"
    }
}

# verify a Splunk access token
function verify_access_token([string]$access_token="", [string]$ingest_url=$INGEST_URL, [bool]$insecure=$INSECURE) {
    if ($insecure) {
        # turn off certificate validation
        [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true} ;
    }
    $url = "$ingest_url/v2/event"
    echo $url
    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        $resp = Invoke-WebRequest -Uri $url -Method POST -ContentType "application/json" -Headers @{"X-Sf-Token"="$access_token"} -Body "[]" -UseBasicParsing
    } catch {
        $err = $_.Exception.Message
        $message = "
        An error occurred while validating the access token
        $err
        "
        throw "$message"
    }
    if (!($resp.StatusCode -Eq 200)) {
        return $false
    } else {
        return $true
    }
}

# create the temp directory if it doesn't exist
function create_temp_dir($tempdir=$tempdir) {
    if ((Test-Path -Path "$tempdir")) {
        Remove-Item -Recurse -Force "$tempdir"
    }
    mkdir "$tempdir" -ErrorAction Ignore
}

# whether the service is running
function service_running([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).State -Eq "Running")
}

# whether the service is installed
function service_installed([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).Name -Eq "$name")
}

# start the service if it's stopped
function start_service([string]$name, [string]$config_path=$config_path) {
    if (!(service_running -name "$name")) {
        if (Test-Path -Path $config_path) {
            try {
                Start-Service -Name "$name"
            } catch {
                $err = $_.Exception.Message
                $message = "
                An error occurred while trying to start the $name service
                $err
                "
                throw "$message"
            }

            # wait for the service to start
            $startTime = Get-Date
            while (!(service_running -name "$name")) {
                # timeout after 30 seconds
                if ((New-TimeSpan -Start $startTime -End (Get-Date)).TotalSeconds -gt 60){
                    throw "The $name service is not running.  Something went wrong during the installation.  Please check the Windows Event Viewer and rerun the installer if necessary."
                }
                # give windows a second to synchronize service status
                Start-Sleep -Seconds 1
            }
        } else {
            throw "$config_path does not exist and is required to start the $name service"
        }
    }
}

# stop the service if it's running
function stop_service([string]$name) {
    if ((service_running -name "$name")) {
        try {
            Stop-Service -Name "$name"
        } catch {
            $err = $_.Exception.Message
            $message = "
            An error occurred while trying to stop the $name service
            $message
            "
        }
    }
}

# download collector package from repo
function download_collector_package([string]$collector_version=$collector_version, [string]$tempdir=$tempdir, [string]$stage=$stage, [string]$arch=$arch, [string]$format=$format) {
    # get the filename to download
    $filename = get_filename -tag $collector_version -format $format -arch $arch
    echo $filename

    # get url for file to download
    $fileurl = get_url -stage $stage -format $format -filename $filename
    echo "Downloading package..."
    download_file -url $fileurl -outputDir $tempdir -filename $filename
    ensure_file_exists "$tempdir\$filename"
    echo "- $fileurl -> '$tempdir'"
}

# check registry for the agent msi package
function msi_installed([string]$name="Splunk OpenTelemetry Collector") {
    return (Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* | Where { $_.DisplayName -eq $name }) -ne $null
}

function update_registry([string]$path, [string]$name, [string]$value) {
    echo "Updating $path for $name..."
    Set-ItemProperty -path "$path" -name "$name" -value "$value"
}

$ErrorActionPreference = 'Stop'; # stop on all errors

# check administrator status
echo 'Checking if running as Administrator...'
if (!(check_if_admin)) {
    throw 'You are not currently running this installation under an Administrator account.  Installation aborted!'
} else {
    echo '- Running as Administrator'
}

# check execution policy
echo 'Checking execution policy'
check_policy

if (msi_installed) {
    throw "The Splunk OpenTelemetry Collector is already installed. Remove/Uninstall the collector and re-run this script."
}

if (service_installed -name "$service_name") {
    throw "The $service_name service is already installed. Remove/Uninstall the collector and re-run this script."
}

if ($with_fluentd -And (service_installed -name "$fluentd_service_name")) {
    throw "The $fluentd_service_name service is already installed. Remove/Uninstall fluentd and re-run this script."
}

if ($with_fluentd -And (Test-Path -Path "$fluentd_base_dir\bin\fluentd")) {
    throw "$fluentd_base_dir\bin\fluentd is already installed. Remove/Uninstall fluentd and re-run this script."
}

if ($ingest_url -eq "") {
    $ingest_url = "https://ingest.$realm.signalfx.com"
}

if ($api_url -eq "") {
    $api_url = "https://api.$realm.signalfx.com"
}

if ($trace_url -eq "") {
    $trace_url = "$ingest_url/v2/trace"
}

if ($hec_url -eq "") {
    $hec_url = "$ingest_url/v1/log"
}

if ($hec_token -eq "") {
    $hec_token = "$access_token"
}

if ($bundle_dir -eq "") {
    $bundle_dir = "C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle"
}

if ("$env:VERIFY_ACCESS_TOKEN" -ne "false") {
    # verify access token
    echo 'Verifying Access Token...'
    if (!(verify_access_token -access_token $access_token -ingest_url $ingest_url -insecure $insecure)) {
        throw "Failed to authenticate access token please verify that your access token is correct"
    }
    else {
        echo '- Verified Access Token'
    }
}

# set up a temporary directory
$tempdir = create_temp_dir -tempdir $tempdir

if ($msi_path -Eq "") {
    # determine package version to fetch
    if ($collector_version -Eq "") {
        echo 'Determining latest release...'
        $collector_version = get_latest -stage $stage -format $format
        echo "- Latest release is $collector_version"
    }

    # download the collector package with the specified collector_version or latest
    download_collector_package -collector_version $collector_version -tempdir $tempdir -stage $stage -arch $arch -format $format

    $msi_path = get_filename -tag $collector_version -format $format -arch $arch
    $msi_path = (Join-Path "$tempdir" "$msi_path")
} else {
    $msi_path = Resolve-Path "$msi_path"
    if (!(Test-Path -Path "$msi_path")) {
        throw "$msi_path not found!"
    }
}

echo "Installing $msi_path ..."
Start-Process msiexec.exe -Wait -ArgumentList "/qn /norestart /i `"$msi_path`""
echo "- Done"

# copy the default configs to $program_data_path
mkdir "$program_data_path" -ErrorAction Ignore
if (!(Test-Path -Path "$agent_config_path") -And (Test-Path -Path "$installation_path\Splunk\OpenTelemetry Collector\agent_config.yaml")) {
    echo "$agent_config_path not found"
    echo "Copying default agent_config.yaml to $agent_config_path"
    Copy-Item "$installation_path\Splunk\OpenTelemetry Collector\agent_config.yaml" "$agent_config_path"
}
if (!(Test-Path -Path "$gateway_config_path") -And (Test-Path -Path "$installation_path\Splunk\OpenTelemetry Collector\gateway_config.yaml")) {
    echo "$gateway_config_path not found"
    echo "Copying default gateway_config.yaml to $gateway_config_path"
    Copy-Item "$installation_path\Splunk\OpenTelemetry Collector\gateway_config.yaml" "$gateway_config_path"
}
if (!(Test-Path -Path "$old_config_path") -And (Test-Path -Path "$installation_path\Splunk\OpenTelemetry Collector\config.yaml")) {
    echo "$old_config_path not found"
    echo "Copying default config.yaml to $old_config_path"
    Copy-Item "$installation_path\Splunk\OpenTelemetry Collector\config.yaml" "$old_config_path"
}

if (($mode -Eq "agent") -And (Test-Path -Path "$agent_config_path")) {
    $config_path = $agent_config_path
} elseif (($mode -Eq "gateway") -And (Test-Path -Path "$gateway_config_path")) {
    $config_path = $gateway_config_path
}

if ($config_path -Eq "") {
    if (Test-Path -Path "$old_config_path") {
        $config_path = $old_config_path
    } else {
        throw "The installed splunk-otel-collector package does not include a supported config file!"
    }
}

update_registry -path "$regkey" -name "SPLUNK_ACCESS_TOKEN" -value "$access_token"
update_registry -path "$regkey" -name "SPLUNK_API_URL" -value "$api_url"
update_registry -path "$regkey" -name "SPLUNK_BUNDLE_DIR" -value "$bundle_dir"
update_registry -path "$regkey" -name "SPLUNK_CONFIG" -value "$config_path"
update_registry -path "$regkey" -name "SPLUNK_HEC_TOKEN" -value "$hec_token"
update_registry -path "$regkey" -name "SPLUNK_HEC_URL" -value "$hec_url"
update_registry -path "$regkey" -name "SPLUNK_INGEST_URL" -value "$ingest_url"
update_registry -path "$regkey" -name "SPLUNK_MEMORY_TOTAL_MIB" -value "$memory"
update_registry -path "$regkey" -name "SPLUNK_REALM" -value "$realm"
update_registry -path "$regkey" -name "SPLUNK_TRACE_URL" -value "$trace_url"

echo "Starting $service_name service..."
start_service -name "$service_name" -config_path "$config_path"
echo "- Started"

if ($with_fluentd) {
    $default_fluentd_config = "$installation_path\Splunk\OpenTelemetry Collector\fluentd\td-agent.conf"
    $default_confd_dir = "$installation_path\Splunk\OpenTelemetry Collector\fluentd\conf.d"

    # copy the default fluentd config to $fluentd_config_path if it does not already exist
    if (!(Test-Path -Path "$fluentd_config_path") -And (Test-Path -Path "$default_fluentd_config")) {
        $default_fluentd_config = Resolve-Path "$default_fluentd_config"
        echo "Copying $default_fluentd_config to $fluentd_config_path"
        mkdir "$fluentd_config_dir" -ErrorAction Ignore | Out-Null
        Copy-Item "$default_fluentd_config" "$fluentd_config_path"
    }

    # copy the default source configs to $fluentd_config_dir\conf.d if it does not already exist
    if (Test-Path -Path "$default_confd_dir\*.conf") {
        mkdir "$fluentd_config_dir\conf.d" -ErrorAction Ignore | Out-Null
        $confFiles = (Get-Item "$default_confd_dir\*.conf")
        foreach ($confFile in $confFiles) {
            $name = $confFile.Name
            $path = $confFile.FullName
            if (!(Test-Path -Path "$fluentd_config_dir\conf.d\$name")) {
                echo "Copying $path to $fluentd_config_dir\conf.d\$name"
                Copy-Item "$path" "$fluentd_config_dir\conf.d\$name"
            }
        }
    }

    echo "Downloading $fluentd_dl_url..."
    download_file -url "$fluentd_dl_url" -outputDir "$tempdir" -fileName "$fluentd_msi_name"
    $fluentd_msi_path = (Join-Path "$tempdir" "$fluentd_msi_name")

    echo "Installing $fluentd_msi_path ..."
    Start-Process msiexec.exe -Wait -ArgumentList "/qn /norestart /i `"$fluentd_msi_path`""
    echo "- Done"

    stop_service -name "$fluentd_service_name"

    echo "Starting $fluentd_service_name service..."
    start_service -name "$fluentd_service_name" -config_path "$fluentd_config_path"
    echo "- Started"
}

# remove the temporary directory
Remove-Item -Recurse -Force "$tempdir"

$message = "
The Splunk OpenTelemetry Connector for Windows has been successfully installed.
Make sure that your system's time is relatively accurate or else datapoints may not be accepted.
The collector's main configuration file is located at $config_path,
and the environment variables are stored in the $regkey registry key.

If the $config_path or any of the SPLUNK_* environment variables in $regkey are modified,
the collector service must be restarted to apply the changes by restarting the system or running the
following PowerShell commands:
  PS> Stop-Service $service_name
  PS> Start-Service $service_name
"
echo "$message"

if ($with_fluentd) {
    $message = "
Fluentd has been installed and configured to forward log events to the Splunk OpenTelemetry Collector.
By default, all log events with the @SPLUNK label will be forwarded to the collector.

The main fluentd configuration file is located at $fluentd_config_path.
Custom input sources and configurations can be added to the $fluentd_config_dir\conf.d directory.
All files with the .conf extension in this directory will automatically be included by fluentd.

By default, fluentd has been configured to collect from the Windows Event Log.
See $fluentd_config_dir\conf.d\eventlog.conf for the default source configuration.

If the fluentd configuration is modified or new config files are added, the fluentd service must be
restarted to apply the changes by restarting the system or running the following PowerShell commands:
  PS> Stop-Service $fluentd_service_name
  PS> Start-Service $fluentd_service_name
"
    echo "$message"
}
