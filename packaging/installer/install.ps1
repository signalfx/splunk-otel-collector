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
    (OPTIONAL) The Splunk realm to use (default: "us0"). The ingest, API, and HEC endpoint URLs will automatically be inferred by this value.
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
.PARAMETER network_interface
    (OPTIONAL) The network interface the collector receivers listen on. (default: "127.0.0.1" for agent mode and "0.0.0.0" otherwise)
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -network_interface "127.0.0.1"
.PARAMETER ingest_url
    (OPTIONAL) Set the base ingest URL explicitly instead of the URL inferred from the specified realm (default: https://ingest.REALM.signalfx.com).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -ingest_url "https://ingest.us1.signalfx.com"
.PARAMETER api_url
    (OPTIONAL) Set the base API URL explicitly instead of the URL inferred from the specified realm (default: https://api.REALM.signalfx.com).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -api_url "https://api.us1.signalfx.com"
.PARAMETER hec_url
    (OPTIONAL) Set the HEC endpoint URL explicitly instead of the endpoint inferred from the specified realm (default: "").
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -hec_url "https://http-inputs-acme.splunkcloud.com/services/collector"
.PARAMETER hec_token
    (OPTIONAL) Set the HEC token if different than the specified Splunk access_token.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -hec_token "HECTOKEN"
.PARAMETER godebug
    (OPTIONAL) Set values for the GODEBUG environment variable.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -godebug "fips140=on"
.PARAMETER with_dotnet_instrumentation
    (OPTIONAL) Whether to install and configure the Splunk Distribution of OpenTelemetry .NET to forward .NET application telemetry to the local collector (default: $false).
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -with_dotnet_instrumentation $true
.PARAMETER deployment_env
    (OPTIONAL) A system-wide "deployment.environment" set via the environment variable 'OTEL_RESOURCE_ATTRIBUTES' for the whole machine. Ignored if -with_dotnet_instrumentation is false.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -with_dotnet_instrumentation $true -deployment_env staging
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
.PARAMETER collector_msi_url
    (OPTIONAL) Specify the URL to the Splunk OpenTelemetry Collector MSI package to install (default: "https://dl.signalfx.com/splunk-otel-collector/msi/release/splunk-otel-collector-<version>-amd64.msi")
    If specified, the -collector_version and -stage parameters will be ignored.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -collector_msi_url https://my.host/splunk-otel-collector-1.2.3-amd64.msi
.PARAMETER msi_path
    (OPTIONAL) Specify a local path to a Splunk OpenTelemetry Collector MSI package to install instead of downloading the package.
    If specified, the -collector_version and -stage parameters will be ignored.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -msi_path "C:\SOME_FOLDER\splunk-otel-collector-1.2.3-amd64.msi"
.PARAMETER dotnet_psm1_path
    (OPTIONAL) Specify a local path to a Splunk OpenTelemetry .NET Auto Instrumentation Powershell Module file (.psm1) instead of downloading the package. This module will be used to install the .NET auto instrumentation files. The most current PSM1 file can be downloaded at https://github.com/signalfx/splunk-otel-dotnet/releases
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -dotnet_psm1_path "C:\SOME_FOLDER\Splunk.OTel.DotNet.psm1"
.PARAMETER dotnet_auto_zip_path
    (OPTIONAL) Specify a local path to a Splunk OpenTelemetry .NET Auto Instrumentation zip package that will be installed by the dotnet psm1 module instead of downloading the package.  The most current zip file can be downloaded at https://github.com/signalfx/splunk-otel-dotnet/releases
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -dotnet_auto_zip_path "C:\SOME_FOLDER\splunk-otel-dotnet-1.2.3-amd64.zip"
.PARAMETER force_skip_verify_access_token
    (OPTIONAL) Forces the skipping the verification check of the Splunk Observability Access Token regardless of what is in the env variable VERIFY_ACCESS_TOKEN.  This is helpful on new installs where access might be an issue or the token isn't created yet.
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -force_skip_verify_access_token $true
.PARAMETER msi_public_properties
    (OPTIONAL) Specify public MSI properties to be used when installing the Splunk OpenTelemetry Collector MSI package.
    For information about the public MSI properties see https://learn.microsoft.com/en-us/windows/win32/msi/property-reference#configuration-properties
    .EXAMPLE
    .\install.ps1 -access_token "ACCESSTOKEN" -msi_public_properties "ARPCOMMENTS=DO_NOT_UNINSTALL" 
.PARAMETER config_path
    (OPTIONAL) Specify a local path to an alternative configuration file for the Splunk OpenTelemetry Collector.
    If specified, the -mode parameter will be ignored.
    .EXAMPLE
    .\install.ps1 -config_path "C:\SOME_FOLDER\my_config.yaml"
.PARAMETER preserve_prev_default_config
   (OPTIONAL) Preserve the default configuration files, located at `$Env:ProgramData\Splunk\OpenTelemetry Collector`, of previous version when upgrading the collector. By default it is $false since version changes can include breaking configuration changes.
   .EXAMPLE
    .\install.ps1 -preserve_prev_default_config $true
#>

param (
    [parameter(Mandatory=$true)][string]$access_token = "",
    [string]$realm = "us0",
    [string]$memory = "512",
    [ValidateSet('agent','gateway')][string]$mode = "agent",
    [string]$network_interface = "",
    [string]$ingest_url = "",
    [string]$api_url = "",
    [string]$hec_url = "",
    [string]$hec_token = "",
    [string]$godebug= "",
    [bool]$insecure = $false,
    [string]$collector_version = "",
    [bool]$with_dotnet_instrumentation = $false,
    [string]$bundle_dir = "",
    [ValidateSet('test','beta','release')][string]$stage = "release",
    [string]$msi_path = "",
    [string]$msi_public_properties = "",
    [string]$config_path = "",
    [bool]$preserve_prev_default_config = $false,
    [string]$collector_msi_url = "",
    [string]$dotnet_psm1_path = "",
    [string]$dotnet_auto_zip_path = "",
    [bool]$force_skip_verify_access_token = $false,
    [string]$deployment_env = "",
    [bool]$UNIT_TEST = $false
)

New-Variable -Name UninstallWildcardRegPath  -Option Constant -Value "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*"
New-Variable -Name CollectorServiceDisplayName -Option Constant -Value "Splunk OpenTelemetry Collector"
$arch = "amd64"
$format = "msi"
$service_name = "splunk-otel-collector"
$signalfx_dl = "https://dl.signalfx.com"
try {
    Resolve-Path $env:PROGRAMFILES 2>&1>$null
    $installation_path = "${env:PROGRAMFILES}\Splunk\OpenTelemetry Collector"
} catch {
    $installation_path = "\Program Files\Splunk\OpenTelemetry Collector"
}
try {
    Resolve-Path $env:PROGRAMDATA 2>&1>$null
    $program_data_path = "${env:PROGRAMDATA}\Splunk\OpenTelemetry Collector"
} catch {
    $program_data_path = "\ProgramData\Splunk\OpenTelemetry Collector"
}
$old_config_path = "$program_data_path\config.yaml"
$agent_config_path = "$program_data_path\agent_config.yaml"
$gateway_config_path = "$program_data_path\gateway_config.yaml"

try {
    Resolve-Path $env:TEMP 2>&1>$null
    $tempdir = "${env:TEMP}\Splunk\OpenTelemetry Collector"
} catch {
    $tempdir = "\tmp\Splunk\OpenTelemetry Collector"
}

# check that we're not running with a restricted execution policy
function check_policy() {
    $executionPolicy  = (Get-ExecutionPolicy)
    $executionRestricted = ($executionPolicy -eq "Restricted")
    if ($executionRestricted) {
        throw @"
You can't import or run scripts with execution policy $executionPolicy.
Change your execution policy to RemoteSigned or similar:
        PS> Set-ExecutionPolicy RemoteSigned
For more information, run the following command:
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
        Your access token could not be verified. This may be due to a network connectivity issue or an invalid access token.
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

function get_service_log_path([string]$name) {
    return "the Windows Event Viewer"
}

# start the service if it's not already running
function start_service([string]$name, [string]$config_path=$null, [int]$timeout=60) {
    $svc = Get-Service -Name $name
    if ($svc.Status -eq "Running") {
        return
    }

    if (!([string]::IsNullOrEmpty($config_path)) -And !(Test-Path -Path $config_path)) {
        throw "$config_path does not exist and is required to start the $name service"
    }

    try {
        if ($svc.Status -ne "ContinuePending" -And $svc.Status -ne "StartPending") {
            $svc.Start()
        }
        $svc.WaitForStatus("Running", [TimeSpan]::FromSeconds($timeout))
    } catch {
        $err = $_.Exception.Message
        $log_path = get_service_log_path -name "$name"
        Write-Warning "An error occurred while trying to start the $name service:"
        Write-Warning "$err"
        Write-Warning "Please check $log_path for more details."
        throw "$err"
    }
}

# stop the service
function stop_service([string]$name, [int]$max_attempts=3, [int]$timeout=60) {
    $svc = Get-Service -Name "$name"
    if ($svc.Status -eq "Stopped") {
        return
    }

    try {
        $svc.Stop()
        $svc.WaitForStatus("Stopped", [TimeSpan]::FromSeconds($timeout))
    } catch {
        $err = $_.Exception.Message
        $log_path = get_service_log_path -name "$name"
        Write-Warning "An error occurred while trying to stop the $name service:"
        Write-Warning "$err"
        Write-Warning "Please check $log_path for more details."
        throw "$err"
    }
}

# download collector package from repo
function download_collector_package([string]$collector_version=$collector_version, [string]$tempdir=$tempdir, [string]$stage=$stage, [string]$arch=$arch, [string]$format=$format) {
    # get the filename to download
    $filename = get_filename -tag $collector_version -format $format -arch $arch

    # get url for file to download
    $fileurl = get_url -stage $stage -format $format -filename $filename
    echo "Downloading $fileName ..."
    download_file -url $fileurl -outputDir $tempdir -filename $filename
    ensure_file_exists "$tempdir\$filename"
    echo "- $fileurl -> '$tempdir'"
}

# check registry for the agent msi package
function is_msi_installed([string]$product_name) {
    return $null -ne (Get-ItemProperty $UninstallWildcardRegPath | Where { $_.DisplayName -eq $product_name })
}

function get_msi_installation_sids([string]$product_name) {
    $sids = [string[]]@()

    $uninstallEntry = Get-ItemProperty $UninstallWildcardRegPath -ErrorAction SilentlyContinue | 
        Where-Object { $_.DisplayName -eq $product_name }
    if ($uninstallEntry) {
        $userInstalls = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Installer\UserData\*\Products\*\InstallProperties' -ErrorAction SilentlyContinue |
            Where-Object { $_.DisplayName -eq $product_name }
        foreach ($user in $userInstalls) {
            # Not all entries are valid user SIDS, e.g.: some are SIDs with suffixes like "_Classes"
            # We only want the SIDs.
            if ($user.PSPath -match 'UserData\\(?<SID>S-1-[0-9\-]+)') {
                $sid = $Matches['SID']
                $sids += , @($sid)
            }
        }
    }

    return $sids
}

function update_registry([string]$path, [string]$name, [string]$value) {
    echo "Updating $path for $name..."
    Set-ItemProperty -path "$path" -name "$name" -value "$value"
}

function set_service_environment([string]$service_name, [hashtable]$env_vars) {
    # Transform the $env_vars to an array of strings so the Set-ItemProperty correctly create the
    # 'Environment' REG_MULTI_SZ value.
    [string []] $multi_sz_value = ($env_vars.Keys | ForEach-Object { "$_=$($env_vars[$_])" } | Sort-Object)

    $target_service_reg_key = Join-Path "HKLM:\SYSTEM\CurrentControlSet\Services" $service_name
    if (Test-Path $target_service_reg_key) {
        Set-ItemProperty $target_service_reg_key -Name "Environment" -Value $multi_sz_value
    }
    else {
        throw "Invalid service '$service_name'. Registry key '$target_service_reg_key' doesn't exist."
    }
}

function install_msi([string]$path) {
    Write-Host "Installing $path ..."
    $startTime = Get-Date
    $proc = (Start-Process msiexec.exe -Wait -PassThru -ArgumentList "/i `"$path`" /qn /norestart $msi_public_properties")
    if ($proc.ExitCode -ne 0 -and $proc.ExitCode -ne 3010) {
        Write-Warning "The installer failed with error code $($proc.ExitCode)."
        try {
            $events = (Get-WinEvent -ProviderName "MsiInstaller" | Where-Object { $_.TimeCreated -ge $startTime })
            ForEach ($event in $events) {
                ($event | Select -ExpandProperty Message | Out-String).TrimEnd() | Write-Host
            }
        } catch {
            Write-Warning "Please check the Windows Event Viewer for more details."
            continue
        }
        Exit $proc.ExitCode
    }
    Write-Host "- Done"
}

function uninstall_msi([string]$product_name) {
    Write-Host "Uninstalling $product_name ..."
    $uninstall_entry = Get-ItemProperty $UninstallWildcardRegPath -ErrorAction SilentlyContinue | 
        Where-Object { $_.DisplayName -eq $product_name } | Select-Object -First 1
    if (-not $uninstall_entry) {
        throw "Failed to find the uninstall registry entry for $product_name"
    }
    $proc = (Start-Process msiexec.exe -Wait -PassThru -ArgumentList "/X `"$($uninstall_entry.PSChildName)`" /qn /norestart")
    if ($proc.ExitCode -ne 0) {
        Write-Warning "The uninstall attempt failed with error code $($proc.ExitCode)."
        Exit $proc.ExitCode
    }
    Write-Host "- Done"
}

$ErrorActionPreference = 'Stop'; # stop on all errors

# check administrator status
echo 'Checking if running as Administrator...'
if (!(check_if_admin)) {
    throw 'Installer is running without Administrator rights. Installation failed.'
} else {
    echo '- Running as Administrator'
}

# check execution policy
echo 'Checking execution policy'
check_policy

if (Get-Service -Name $service_name -ErrorAction SilentlyContinue) {
    Write-Host "The $service_name service is already installed. Checking installation for automatic update."

    $uninstall_collector = $true
    $collector_sids = get_msi_installation_sids -product_name $CollectorServiceDisplayName
    if ($collector_sids.Count -eq 0) {
        $uninstall_collector = $false
        Write-Warning "The $service_name service exists but it is not on the Windows installation database."
    }
    else {
        if ($collector_sids.Count -gt 1) {
            $sids_list = $collector_sids -join ", "
            throw "The $CollectorServiceDisplayName is already installed for multiple users (SIDs: $sids_list). Uninstall the collector and remove remaining users installations from the registry."
        }

        # "S-1-5-18" is the SID for the Local System account, which is used for machine-wide installations.
        if ("S-1-5-18" -ne $collector_sids[0]) {
            # not a machine wide installation, check if it is the same user
            $currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent()
            $currentUserSID = $currentUser.User.Value
            if ($currentUserSID -ne $collector_sids[0]) {
                $sid = New-Object System.Security.Principal.SecurityIdentifier($userSid)
                $user = $sid.Translate([System.Security.Principal.NTAccount])
                throw "The $CollectorServiceDisplayName was last installed by '${user.Value}' it must be updated or uninstalled by the same user." 
            }
        }
    }

    Write-Host "Stopping $service_name service..."
    stop_service -name "$service_name"
    if ($uninstall_collector) {
        uninstall_msi -product_name $CollectorServiceDisplayName
    }
    if (-not $preserve_prev_default_config) {
        $default_config_files = @("agent_config.yaml", "gateway_config.yaml")
        foreach ($file in $default_config_files) {
            $target = Join-Path "${Env:ProgramData}\Splunk\OpenTelemetry Collector" "$file"
            Write-Host "Deleting previous version default configuration file '$target'"
            Remove-Item -Path $target
        }
    }
}

# create a temporary directory
$tempdir = create_temp_dir -tempdir $tempdir

if ($with_dotnet_instrumentation) {
    if ((is_msi_installed -name "SignalFx .NET Tracing 64-bit") -Or (is_msi_installed -name "SignalFx .NET Tracing 32-bit")) {
        throw "SignalFx .NET Instrumentation is already installed. Stop all instrumented applications and uninstall SignalFx Instrumentation for .NET before running this script again."
    }
    echo "Downloading Splunk Distribution of OpenTelemetry .NET ..."
    if ($dotnet_psm1_path -eq "") {
        $module_name = "Splunk.OTel.DotNet.psm1"
        $download = "https://github.com/signalfx/splunk-otel-dotnet/releases/latest/download/$module_name"
        $dotnet_autoinstr_path = Join-Path $tempdir $module_name
        echo "Downloading .NET Instrumentation installer ..."
        try {
            [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
            Invoke-WebRequest -Uri $download -OutFile $dotnet_autoinstr_path -UseBasicParsing
        } catch {
            $err = $_.Exception.Message
            $message = "
            An error occured when trying to download .NET Instrumentation installer from $download. This may be due to a network connectivity issue.
            $err
            "
            throw "$message"
        }
        Import-Module $dotnet_autoinstr_path
    } else {
        $dotnet_autoinstr_path = $dotnet_psm1_path
        echo "Using Local PSM1 file and ArgumentList values: $dotnet_psm1_path -ArgumentList $dotnet_auto_zip_path"
        Import-Module $dotnet_autoinstr_path -ArgumentList $dotnet_auto_zip_path
    }
    
}

if ($ingest_url -eq "") {
    $ingest_url = "https://ingest.$realm.signalfx.com"
}

if ($api_url -eq "") {
    $api_url = "https://api.$realm.signalfx.com"
}

if ($bundle_dir -eq "") {
    $bundle_dir = "$installation_path\agent-bundle"
}

if ($force_skip_verify_access_token) {
    echo 'Skipping Access Token verification'
} else {   
    if ("$env:VERIFY_ACCESS_TOKEN" -ne "false") {
        # verify access token
        echo 'Verifying Access Token...'
        if (!(verify_access_token -access_token $access_token -ingest_url $ingest_url -insecure $insecure)) {
            throw "Access token authentication failed. Verify that your access token is correct."
        }
        else {
            echo '- Verified Access Token'
        }
    }
}

if ($collector_msi_url) {
    $collector_msi_name = "splunk-otel-collector.msi"
    echo "Downloading $collector_msi_url..."
    download_file -url "$collector_msi_url" -outputDir "$tempdir" -fileName "$collector_msi_name"
    $msi_path = (Join-Path "$tempdir" "$collector_msi_name")
} elseif ($msi_path -Eq "") {
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

install_msi -path "$msi_path"

# copy the default configs to $program_data_path
mkdir "$program_data_path" -ErrorAction Ignore
if (!(Test-Path -Path "$agent_config_path") -And (Test-Path -Path "$installation_path\agent_config.yaml")) {
    echo "$agent_config_path not found"
    echo "Copying default agent_config.yaml to $agent_config_path"
    Copy-Item "$installation_path\agent_config.yaml" "$agent_config_path"
}
if (!(Test-Path -Path "$gateway_config_path") -And (Test-Path -Path "$installation_path\gateway_config.yaml")) {
    echo "$gateway_config_path not found"
    echo "Copying default gateway_config.yaml to $gateway_config_path"
    Copy-Item "$installation_path\gateway_config.yaml" "$gateway_config_path"
}
if (!(Test-Path -Path "$old_config_path") -And (Test-Path -Path "$installation_path\config.yaml")) {
    echo "$old_config_path not found"
    echo "Copying default config.yaml to $old_config_path"
    Copy-Item "$installation_path\config.yaml" "$old_config_path"
}

if ($config_path -Eq "") {
    if (($mode -Eq "agent") -And (Test-Path -Path "$agent_config_path")) {
        $config_path = $agent_config_path
    } elseif (($mode -Eq "gateway") -And (Test-Path -Path "$gateway_config_path")) {
        $config_path = $gateway_config_path
    } elseif (Test-Path -Path "$old_config_path") {
        $config_path = $old_config_path
    }
}

if (!(Test-Path -Path "$config_path")) {
    throw "Valid Collector configuration file not found at $config_path."
}

$collector_env_vars = @{
    "SPLUNK_ACCESS_TOKEN"     = "$access_token";
    "SPLUNK_API_URL"          = "$api_url";
    "SPLUNK_BUNDLE_DIR"       = "$bundle_dir";
    "SPLUNK_CONFIG"           = "$config_path";
    "SPLUNK_HEC_TOKEN"        = "$hec_token";
    "SPLUNK_HEC_URL"          = "$hec_url";
    "SPLUNK_INGEST_URL"       = "$ingest_url";
    "SPLUNK_MEMORY_TOTAL_MIB" = "$memory";
    "SPLUNK_REALM"            = "$realm";
}

if ($network_interface -Ne "") {
    $collector_env_vars.Add("SPLUNK_LISTEN_INTERFACE", "$network_interface")
}

if ($godebug -Ne "") {
    $collector_env_vars.Add("GODEBUG", "$godebug")
}

# set the environment variables for the collector service
set_service_environment $service_name $collector_env_vars

$message = "
The $CollectorServiceDisplayName for Windows has been successfully installed.
Make sure that your system's time is relatively accurate or else datapoints may not be accepted.
The collector's main configuration file is located at $config_path,
and the environment variables are stored in the $regkey registry key.

If the $config_path configuration file or any of the
SPLUNK_* environment variables in the $regkey registry key are modified,
the collector service must be restarted to apply the changes by restarting the system or running the
following PowerShell commands:
  PS> Stop-Service $service_name
  PS> Start-Service $service_name
"
echo "$message"

$otel_resource_attributes = ""
if ($deployment_env -ne "") {
    echo "Setting deployment environment to $deployment_env"
    $otel_resource_attributes = "deployment.environment=$deployment_env"
} else {
    echo "Deployment environment was not specified. Unless otherwise defined, will appear as 'unknown' in the UI."
}

if ($with_dotnet_instrumentation) {
    echo "Installing Splunk Distribution of OpenTelemetry .NET..."
    $currentInstallVersion = Get-OpenTelemetryInstallVersion
    if ($currentInstallVersion) {
        throw "The Splunk Distribution of OpenTelemetry .NET is already installed. Stop all instrumented applications and uninstall it and then rerun this script."
    }

    # If the variable dotnet_auto_zip_path is an empty string, then the Installer will download the .NET Instrumentation from the default repository.
    Install-OpenTelemetryCore -LocalPath $dotnet_auto_zip_path

    $installed_version = Get-OpenTelemetryInstallVersion
    if ($otel_resource_attributes -ne "") {
        $otel_resource_attributes += ","
    }
    $otel_resource_attributes += "splunk.zc.method=splunk-otel-dotnet-$installed_version"
}

if ($otel_resource_attributes -ne "") {
    # The OTEL_RESOURCE_ATTRIBUTES environment variable must be set before restarting IIS.
    $regkey = "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment"
    try {
        update_registry -path "$regkey" -name "OTEL_RESOURCE_ATTRIBUTES" -value "$otel_resource_attributes"
    } catch {
        Write-Warning "Failed to set OTEL_RESOURCE_ATTRIBUTES environment variable."
        continue
    }
}

if ($with_dotnet_instrumentation) {
    if (Get-Service -Name "W3SVC" -ErrorAction SilentlyContinue) {
        echo "Registering OpenTelemetry for IIS..."
        Register-OpenTelemetryForIIS
    }

    $message = "
Splunk Distribution of OpenTelemetry for .NET has been installed and configured to forward traces to the $CollectorServiceDisplayName.
By default, the .NET instrumentation will automatically generate telemetry only for .NET applications running on IIS.
"
    echo "$message"
}

# remove the temporary directory
Remove-Item -Recurse -Force "$tempdir"

# Try starting the service(s) only after all components were successfully installed.
echo "Starting $service_name service..."
start_service -name "$service_name" -config_path "$config_path"
echo "- Started"

if (($network_interface -Eq "") -And ($mode -Eq "agent")) {
    echo "[NOTICE] Starting with version 0.86.0, the collector installer changed its default network listening interface from 0.0.0.0 to 127.0.0.1 for agent mode. Please consult the release notes for more information and configuration options."
}
