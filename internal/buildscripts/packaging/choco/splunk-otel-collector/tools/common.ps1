$installation_path = "$drive" + "\Program Files\Splunk\OpenTelemetry Collector"
$program_data_path = "$drive" + "\ProgramData\Splunk\OpenTelemetry Collector"
$config_path = "$program_data_path\"

$service_name = "splunk-otel-collector"
# whether the service is running
function service_running([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = 'splunk-otel-collector'" | Select Name, State).State -Eq "Running")
}

# whether the service is installed
function service_installed([string]$name) {
    return ((Get-CimInstance -ClassName win32_service -Filter "Name = '$name'" | Select Name, State).Name -Eq "$name")
}

# start the service if it's stopped
function start_service([string]$name=$service_name, [string]$config_path=$config_path) {
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
                # timeout after 60 seconds
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
    if (service_running -name "$name") {
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
    echo "Updating $path for $name..."
    Set-ItemProperty -path "$path" -name "$name" -value "$value"
}

# wait for the service to start
function wait_for_service([string]$name=$service_name, [int]$timeout=60) {
    $startTime = Get-Date
    while (!(service_running -name "$name")){
        if ((New-TimeSpan -Start $startTime -End (Get-Date)).TotalSeconds -gt $timeout){
            throw "Service is not running.  Something went wrong durring the installation.  Please rerun the installer"
        }
        # give windows a second to synchronize service status
        Start-Sleep -Seconds 1
    }
}

$ErrorActionPreference = 'Stop'; # stop on all errors
