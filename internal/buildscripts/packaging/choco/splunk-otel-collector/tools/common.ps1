$installation_path = "${env:PROGRAMFILES}\Splunk\OpenTelemetry Collector"
$program_data_path = "${env:PROGRAMDATA}\Splunk\OpenTelemetry Collector"
$config_path = "$program_data_path\"

$service_name = "splunk-otel-collector"

try {
    Resolve-Path $env:SYSTEMDRIVE 2>&1>$null
    $fluentd_base_dir = "${env:SYSTEMDRIVE}\opt\td-agent"
} catch {
    $fluentd_base_dir = "\opt\td-agent"
}
$fluentd_config_dir = "$fluentd_base_dir\etc\td-agent"
$fluentd_config_path = "$fluentd_config_dir\td-agent.conf"
$fluentd_service_name = "fluentdwinsvc"
$fluentd_log_path = "$fluentd_base_dir\td-agent.log"

function get_service_log_path([string]$name) {
    $log_path = "the Windows Event Viewer"
    if (($name -eq $fluentd_service_name) -and (Test-Path -Path "$fluentd_log_path")) {
        $log_path = $fluentd_log_path
    }
    return $log_path
}

# start the service if it's not already running
function start_service([string]$name, [string]$config_path=$null, [int]$timeout=60) {
    $svc = Get-Service -Name $name
    if ($svc.Status -eq "Running") {
        return
    }

    if (!($config_path -eq $null) -And !(Test-Path -Path $config_path)) {
        throw "$config_path does not exist and is required to start the $name service"
    }

    try {
        $svc.Start()
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
function stop_service([string]$name, [int]$timeout=60) {
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

# remove registry entries created by the splunk-otel-collector service
function remove_otel_registry_entries() {
    try {
        if (Test-Path "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\splunk-otel-collector"){
            Remove-Item "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\splunk-otel-collector"
        }
    } catch {
        $err = $_.Exception.Message
        $message = "
        unable to remove registry entries at HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\splunk-otel-collector
        $err
        "
        throw "$message"
    }
}

function set_env_var_value_from_package_params([hashtable] $env_vars, [hashtable] $package_params, [string]$name, [string]$default_value) {
    $value = $package_params[$name]
    if ($value) {
        # If the variable was passed as a package parameter, use that value.
        $env_vars[$name] = $value
        return
    }

    # If the variable was not passed as a package parameter, check if it was already set in the environment.
    $value = $env_vars[$name]
    if ($value) {
        # If the variable already exists in the environment, use that value.
        return
    }

    $value = "$default_value" # Env. var values are always strings.
    $env_vars[$name] = $value
    Write-Host "The $name package parameter was not set, using the default value: '$value'"
}

function set_service_environment([string]$service_name, [hashtable]$env_vars) {
    Write-Host "Setting environment variables for the $service_name service..."
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

# create the temp directory if it doesn't exist
function create_temp_dir($tempdir) {
    if ((Test-Path -Path "$tempdir")) {
        Remove-Item -Recurse -Force "$tempdir"
    }
    mkdir "$tempdir" -ErrorAction Ignore
}

function install_msi([string]$path) {
    Write-Host "Installing $path ..."
    $startTime = Get-Date
    $proc = (Start-Process msiexec.exe -Wait -PassThru -ArgumentList "/qn /norestart /i `"$path`"")
    if ($proc.ExitCode -ne 0 -and $proc.ExitCode -ne 3010) {
        Write-Warning "The installer failed with error code ${proc.ExitCode}."
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

$ErrorActionPreference = 'Stop'; # stop on all errors
