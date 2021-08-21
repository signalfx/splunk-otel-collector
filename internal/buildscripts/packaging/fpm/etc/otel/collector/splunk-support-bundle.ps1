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

#######################################
# Globals
#######################################
$CONFDIR="C:\Program Files\Splunk\OpenTelemetry Collector" # Default configuration directory
$DIRECTORY= # Either passed as CLI parameter or later set to CONFDIR
$TMPDIR="C:\Program Files\Splunk\splunk-support-bundle-$([int64](New-TimeSpan -Start (Get-Date "01/01/1970") -End (Get-Date)).TotalSeconds)" # Unique temporary directory for support bundle contents

$ErrorActionPreference= 'silentlycontinue'

function usage {
    "This is help for this program. It does nothing. Hope that helps."
    write-host "USAGE: [-help] [-d directory] [-t directory]"
    write-host "  -d      directory where Splunk OpenTelemetry Connector configuration is located"
    write-host "          (if not specified, defaults to C:\Program Files\Splunk\OpenTelemetry Collector)"
    write-host "  -t      Unique temporary directory for support bundle contents"
    write-host "          (if not specified, defaults to C:\Program Files\Splunk\OpenTelemetry Collector)"
    write-host "  -help   display help"
    exit 1
}

#######################################
# Parse command line arguments
#######################################
for ( $i = 0; $i -lt $args.count; $i++ ) {
    if (($args[$i] -eq "-d") -OR ($args[$i] -eq "-directory")) {
        if (($args[$i+1]) -AND ($args[$i+1] -ne "-t") -AND ($args[$i+1] -ne "-tempdir") -AND ($args[$i+1] -ne "-h") -AND ($args[$i+1] -ne "-help")) {
            $CONFDIR = $args[$i+1];
            $i = $i + 1;
        } else {
            usage;
        }
    } elseif (($args[$i] -eq "-t") -OR ($args[$i] -eq "-directory")) {
        if (($args[$i+1]) -AND ($args[$i+1] -ne "-d") -AND ($args[$i+1] -ne "-directory") -AND ($args[$i+1] -ne "-h") -AND ($args[$i+1] -ne "-help")) {
            $TMPDIR = $args[$i+1];
            $i = $i + 1;
        } else {
            usage;
        }
    } elseif (($args[$i] -eq "-h") -OR ($args[$i] -eq "-help")) {
        usage;
    } else {
        usage;
    }
}

#######################################
# Creates a unique temporary directory to store the contents of the support
# bundle. Do not attempt to cleanup to prevent any accidental deletions.
# This command can only be run once per second or will error out.
# This script could result in a lot of temporary data if run multiple times.
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0 if successful, non-zero on error.
#######################################
function createTempDir {
    write-host "INFO: Creating temporary directory..."
    if (Test-Path -Path $TMPDIR) {
        write-host "ERROR: TMPDIR ($TMPDIR) exists. Exiting."
        exit 1
    } else {
        New-Item -Path $TMPDIR -ItemType Directory | Out-Null
        New-Item -Path $TMPDIR/logs -ItemType Directory | Out-Null
        New-Item -Path $TMPDIR/metrics -ItemType Directory | Out-Null
        New-Item -Path $TMPDIR/zpages -ItemType Directory | Out-Null
	# We can not create directory using special characters like : , ? 
	# So we have encoded it and then created new directory.
        $global:DIRECTORY = [System.Web.HTTPUtility]::UrlEncode("localhost:55679")
        New-Item -Path $TMPDIR/zpages/$global:DIRECTORY -ItemType Directory | Out-Null
        New-Item -Path $TMPDIR/zpages/$global:DIRECTORY/debug -ItemType Directory | Out-Null
    }
}

#######################################
# Gather configuration
# Without this it is very hard to troubleshoot issues so exit if no permissions.
#  - GLOBALS: CONFDIR, DIRECTORY, TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0 if successful, non-zero on error.
#######################################
function getConfig {
    write-host "INFO: Getting configuration..."
# If directory does not exist the support bundle is useless so exit
    if (-NOT (Test-Path -Path $CONFDIR)) {
        write-host "ERROR: Could not find directory ($CONFDIR)."
        usage
    } else {
        Copy-Item -Path "$CONFDIR" -Destination "$TMPDIR/config" -Recurse
    }
# Also need to get config in memory as dynamic config may modify stored config
# It's possible user has disabled collecting in memory config
    $connection = New-Object System.Net.Sockets.TcpClient("localhost", 55554)
    if ($connection.Connected) {
        cmd.exe /c "curl -s http://localhost:55554/debug/configz/initial > $TMPDIR/config/initial.yaml 2>&1"
        cmd.exe /c "curl -s http://localhost:55554/debug/configz/effective > $TMPDIR/config/effective.yaml 2>&1"
    } else { 
        Write-Host "WARN: localhost:55554 unavailable so in memory configuration not collected"
    }
}

#######################################
# Gather status
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
function getStatus {
    Write-Host "INFO: Getting status..."
    Get-Service splunk-otel-collector > $TMPDIR/logs/splunk-otel-collector.txt 2>&1
    Get-Service fluentdwinsvc > $TMPDIR/logs/td-agent.txt 2>&1
}

#######################################
# Gather logs
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
function getLogs {
    Write-Host "INFO: Getting logs..."
    Get-EventLog -LogName Application -Source "splunk-otel-collector" > $TMPDIR/logs/splunk-otel-collector.log 2>&1
    Get-EventLog -LogName Application -Source "td-agent" > $TMPDIR/logs/splunk-otel-collector.log 2>&1
    $LOGDIR="/var/log/td-agent"
    if (Test-Path -Path $LOGDIR) {
        Copy-Item -Path "$LOGDIR" -Destination "$TMPDIR/logs/td-agent/" -Recurse
    } else {
        Write-Host "WARN: Permission denied to directory ($LOGDIR)."
    }
}

#######################################
# Gather metrics
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
function getMetrics {
    Write-Host "INFO: Getting metric information..."
    $connection = New-Object System.Net.Sockets.TcpClient("localhost", 8888)
    if ($connection.Connected) {
        cmd.exe /c "curl -s http://localhost:8888/metrics > $TMPDIR/metrics/collector-metrics.txt 2>&1"
    } else { 
        Write-Host "WARN: localhost:8888/metrics unavailable so metrics not collected"
    }
}

#######################################
# Gather zpages
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
function getZpages {
    Write-Host "INFO: Getting zpages information..."
    $connection = New-Object System.Net.Sockets.TcpClient("localhost", 55679)
    if ($connection.Connected) {
        cmd.exe /c "curl -s http://localhost:55679/debug/tracez > $TMPDIR/zpages/tracez.html 2>&1"
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        $packages = Invoke-WebRequest -Uri "http://localhost:55679/debug/tracez" -UseBasicParsing
        foreach ($package in $packages.links.href) {
            $ENCODED_PACKAGE_NAME = [System.Web.HTTPUtility]::UrlEncode("$package")
            cmd.exe /c "curl -s `"http://localhost:55679/debug/$package`" > $TMPDIR/zpages/$global:DIRECTORY/debug/$ENCODED_PACKAGE_NAME 2>&1"
        }
    } else { 

        Write-Host "WARN: localhost:55679 unavailable so zpages not collected"
    }

        
}

#######################################
# Gather System information
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
function getHostInfo {
    Write-Host "INFO: Getting host information..."
    for ( $i = 0; $i -lt 3; $i++ ) {
        Get-Process -Name 'otelcol' >> $TMPDIR/metrics/top.txt 2>&1 
        Get-Process -Name 'fluentd' >> $TMPDIR/metrics/top.txt 2>&1 
        Start-Sleep -s 2
    }
    if (-NOT (Get-Process -Name 'otelcol')) {
        Write-Host "WARN: Unable to find otelcol PIDs"
        Write-Host "      top will not be collected for otelcol";
    }
    if (-NOT (Get-Process -Name 'fluentd')) {
        Write-Host "WARN: Unable to find fluentd PIDs"
        Write-Host "      top will not be collected for fluentd";
    }
    
    Get-PSDrive > $TMPDIR/metrics/df.txt 2>&1
    
    Get-CIMInstance Win32_OperatingSystem | Select TotalVisibleMemorySize,FreePhysicalMemory,TotalVirtualMemorySize,FreeVirtualMemory > $TMPDIR/metrics/free.txt 2>&1
}

#######################################
# Tar support bundle
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0 if successful, non-zero on error
#######################################
function tarResults {
    Write-Host "INFO: Creating tarball..."
    $TAR_NAME = Split-Path $TMPDIR -leaf
    tar -cf "$TAR_NAME.tar.gz" $TMPDIR 2>&1
    if (Test-Path -Path "./$TAR_NAME.tar.gz") {
        Write-Host "INFO: Support bundle available at: ./$TAR_NAME.tar.gz"
        Write-Host "      Please attach this to your support case"
        exit 0
    } else {
        Write-Host "ERROR: Support bundle was not properly created."
        Write-Host "       See $TMPDIR/stdout.log for more information."
        exit 1
    }
}

createTempDir
getConfig
getStatus
getLogs
getMetrics
tarResults
getHostInfo

# Attempt to generate a support bundle
# Capture all output
createTempDir
main 2>&1 | tee -a "$TMPDIR"/stdout.log