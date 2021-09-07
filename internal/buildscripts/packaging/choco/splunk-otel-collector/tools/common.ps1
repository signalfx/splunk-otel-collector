try {
    $drive = (Get-ToolsLocation | Split-Path -Qualifier)
} catch {
    $drive = ""
}
$installation_path = "$drive" + "\Program Files\Splunk\OpenTelemetry Collector"
$program_data_path = "$drive" + "\ProgramData\Splunk\OpenTelemetry Collector"
$config_path = "$program_data_path\agent_config.yaml"

function get_value_from_file([string]$path) {
    $value = ""
    if (Test-Path -Path "$path") {
        try {
            $value = (Get-Content -Path "$path").Trim()
        } catch {
            $value = ""
        }
    }
    return "$value"
}

# create directories in program data
function create_program_data() {
    if (!(Test-Path -Path "$program_data_path")) {
        echo "Creating $program_data_path"
        (mkdir "$program_data_path")
    }
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

# check registry for the agent msi package
function msi_installed([string]$name="Splunk OpenTelemetry Collector") {
    return (Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* | Where { $_.DisplayName -eq $name }) -ne $null
}

function update_registry([string]$path, [string]$name, [string]$value) {
    echo "Updating $path for $name..."
    Set-ItemProperty -path "$path" -name "$name" -value "$value"
}

$ErrorActionPreference = 'Stop'; # stop on all errors
