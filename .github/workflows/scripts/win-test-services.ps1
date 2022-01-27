param (
    [string]$mode = "agent",
    [string]$with_fluentd = "true"
)

$ErrorActionPreference = 'Stop'
Set-PSDebug -Trace 1

$SPLUNK_CONFIG = Get-ItemPropertyValue -PATH "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_CONFIG"
$SPLUNK_CONFIG_FILE = Split-Path $SPLUNK_CONFIG -leaf
if ( "$SPLUNK_CONFIG_FILE" -ne "${mode}_config.yaml" ) {
    write-host "Environment variable SPLUNK_CONFIG is not properly set."
    exit 1
}

if ((Get-CimInstance -ClassName win32_service -Filter "Name = 'splunk-otel-collector'" | Select Name, State).State -Eq "Running") {
    write-host "splunk-otel-collector service is running."
} else {
    throw "Failed to install splunk-otel-collector using chocolatey."
}

if ("$with_fluentd" -eq "true") {
    if ((Get-CimInstance -ClassName win32_service -Filter "Name = 'fluentdwinsvc'" | Select Name, State).State -Eq "Running") {
        write-host "fluentdwinsvc service is running."
    } else {
        throw "Failed to install fluentdwinsvc using chocolatey."
    }
} else {
    if ((Get-CimInstance -ClassName win32_service -Filter "Name = 'fluentdwinsvc'" | Select Name, State).State -Eq "Running") {
        throw "fluentdwinsvc service is running."
    } else {
        write-host "fluentdwinsvc service is not running."
    }
}
