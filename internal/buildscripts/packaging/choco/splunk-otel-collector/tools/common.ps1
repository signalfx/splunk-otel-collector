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

# whether the service is running
function service_running([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).State -Eq "Running")
}

# whether the service is installed
function service_installed([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).Name -Eq "$name")
}

function get_service_log_path([string]$name) {
    $log_path = "the Windows Event Viewer"
    if (($name -eq $fluentd_service_name) -and (Test-Path -Path "$fluentd_log_path")) {
        $log_path = $fluentd_log_path
    }
    return $log_path
}

# wait for the service to start
function wait_for_service([string]$name, [int]$timeout=60) {
    $startTime = Get-Date
    while (!(service_running -name "$name")){
        if ((New-TimeSpan -Start $startTime -End (Get-Date)).TotalSeconds -gt $timeout) {
            $err = "Timed out waiting for the $name service to be running."
            $log_path = get_service_log_path -name "$name"
            Write-Warning "$err"
            Write-Warning "Please check $log_path for more details."
            throw "$err"
        }
        # give windows a second to synchronize service status
        Start-Sleep -Seconds 1
    }
}

# wait for the service to stop
function wait_for_service_stop([string]$name, [int]$timeout=60) {
    $startTime = Get-Date
    while ((service_running -name "$name")){
        if ((New-TimeSpan -Start $startTime -End (Get-Date)).TotalSeconds -gt $timeout) {
            $err = "Timed out waiting for the $name service to be stopped."
            $log_path = get_service_log_path -name "$name"
            Write-Warning "$err"
            Write-Warning "Please check $log_path for more details."
            throw "$err"
        }
        # give windows a second to synchronize service status
        Start-Sleep -Seconds 1
    }
}

# start the service if it's stopped
function start_service([string]$name, [string]$config_path=$config_path, [int]$max_attempts=3, [int]$timeout=60) {
    if (!(service_installed -name "$name")) {
        throw "The $name service does not exist!"
    }
    if (!(service_running -name "$name")) {
        if (Test-Path -Path $config_path) {
            for ($i=1; $i -le $max_attempts; $i++) {
                try {
                    Start-Service -Name "$name"
                    break
                } catch {
                    $err = $_.Exception.Message
                    if ($i -eq $max_attempts) {
                        $log_path = get_service_log_path -name "$name"
                        Write-Warning "An error occurred while trying to start the $name service:"
                        Write-Warning "$err"
                        Write-Warning "Please check $log_path for more details."
                        throw "$err"
                    } else {
                        Stop-Service -Name "$name" -ErrorAction Ignore
                        Start-Sleep -Seconds 10
                        continue
                    }
                }
            }
            wait_for_service -name "$name" -timeout $timeout
        } else {
            throw "$config_path does not exist and is required to start the $name service"
        }
    }
}

# stop the service if it's running
function stop_service([string]$name, [int]$max_attempts=3, [int]$timeout=60) {
    if ((service_running -name "$name")) {
        for ($i=1; $i -le $max_attempts; $i++) {
            try {
                Stop-Service -Name "$name"
                break
            } catch {
                $err = $_.Exception.Message
                if ($i -eq $max_attempts) {
                    $log_path = get_service_log_path -name "$name"
                    Write-Warning "An error occurred while trying to start the $name service:"
                    Write-Warning "$err"
                    Write-Warning "Please check $log_path for more details."
                    throw "$err"
                } else {
                    Start-Sleep -Seconds 10
                    continue
                }
            }
        }
        wait_for_service_stop -name "$name" -timeout $timeout
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

function update_registry([string]$path, [string]$name, [string]$value) {
    write-host "Updating $path for $name..."
    Set-ItemProperty -path "$path" -name "$name" -value "$value"
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
