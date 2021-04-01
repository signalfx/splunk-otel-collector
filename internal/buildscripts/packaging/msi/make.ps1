# Copyright 2020 Splunk, Inc.
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

<#
.SYNOPSIS
    Makefile like build commands for the Collector on Windows.

    Usage:   .\make.ps1 <Command> [-<Param> <Value> ...]
    Example: .\make.ps1 New-MSI -Config "./my-config.yaml" -Version "v0.0.2"
.PARAMETER Target
    Build target to run (Install-Tools, New-MSI)
#>
Param(
    [Parameter(Mandatory=$true, ValueFromRemainingArguments=$true)][string]$Target
)

$ErrorActionPreference = "Stop"

function Install-Tools {
    # disable progress bar support as this causes CircleCI to crash
    $OriginalPref = $ProgressPreference
    $ProgressPreference = "SilentlyContinue"
    Install-WindowsFeature Net-Framework-Core
    $ProgressPreference = $OriginalPref

    choco install wixtoolset -y
    setx /m PATH "%PATH%;C:\Program Files (x86)\WiX Toolset v3.11\bin"
    refreshenv
}

function New-MSI(
    [string]$Otelcol="./bin/otelcol_windows_amd64.exe",
    [string]$Version="0.0.1",
    [string]$BuildDir="./dist",
    [string]$Config="./cmd/otelcol/config/collector/agent_config.yaml",
    [string]$FluentdConfig="./internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd/fluent.conf",
    [string]$FluentdConfDir="./internal/buildscripts/packaging/msi/fluentd/conf.d"
) {
    $msiName = "splunk-otel-collector-$Version-amd64.msi"
    $filesDir = "$BuildDir\msi"
    if (Test-Path "$filesDir") {
        Remove-Item -Force -Recurse "$filesDir"
    }
    mkdir "$filesDir\fluentd\conf.d" -ErrorAction Ignore
    Copy-Item "$Config" "$filesDir\config.yaml"
    Copy-Item "$FluentdConfig" "$filesDir\fluentd\td-agent.conf"
    Copy-Item "$FluentdConfDir\*.conf" "$filesDir\fluentd\conf.d" -Recurse
    heat dir "$filesDir" -srd -sreg -gg -template fragment -cg ConfigFiles -dr INSTALLDIR -out "$BuildDir\configfiles.wsx"
    candle -arch x64 -out "$BuildDir\configfiles.wixobj" "$BuildDir\configfiles.wsx"
    candle -arch x64 -out "$BuildDir\splunk-otel-collector.wixobj" -dVersion="$Version" -dOtelcol="$Otelcol" .\internal\buildscripts\packaging\msi\splunk-otel-collector.wxs
    light -ext WixUtilExtension.dll -sval -spdb -out "$BuildDir\$msiName" -b "$filesDir" "$BuildDir\splunk-otel-collector.wixobj" "$BuildDir\configfiles.wixobj"
    if (!(Test-Path "$BuildDir\$msiName")) {
        throw "$BuildDir\$msiName not found!"
    }
}

function Confirm-MSI {
    # ensure system32 is in Path so we can use executables like msiexec & sc
    $env:Path += ";C:\Windows\System32"
    $msipath = Resolve-Path "$pwd\dist\splunk-otel-collector-*-amd64.msi"

    # install msi, validate service is installed & running
    echo "Installing $msipath ..."
    Start-Process -Wait msiexec "/i `"$msipath`" /qn"
    sc.exe query state=all | findstr "splunk-otel-collector" | Out-Null
    if ($LASTEXITCODE -ne 0) { Throw "splunk-otel-collector service failed to install" }

    $configpath = Resolve-Path "\ProgramData\Splunk\OpenTelemetry Collector\config.yaml"
    echo "Updating $configpath ..."
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_ACCESS_TOKEN}', 'testing123') | Set-Content -Path "$configpath"
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_API_URL}', 'https://api.us0.signalfx.com') | Set-Content -Path "$configpath"
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_HEC_URL}', 'https://ingest.us0.signalfx.com/v1/log') | Set-Content -Path "$configpath"
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_HEC_TOKEN}', 'testing456') | Set-Content -Path "$configpath"
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_INGEST_URL}', 'https://ingest.us0.signalfx.com') | Set-Content -Path "$configpath"
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_TRACE_URL}', 'https://ingest.us0.signalfx.com/v2/trace') | Set-Content -Path "$configpath"
    ((Get-Content -path "$configpath" -Raw) -Replace '\${SPLUNK_BUNDLE_DIR}', 'C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle') | Set-Content -Path "$configpath"

    # start service
    echo "Starting service ..."
    Start-Service splunk-otel-collector

    # uninstall msi, validate service is uninstalled
    echo "Uninstalling $msipath ..."
    Start-Process -Wait msiexec "/x `"$msipath`" /qn"
    sc.exe query state=all | findstr "splunk-otel-collector" | Out-Null
    if ($LASTEXITCODE -ne 1) { Throw "splunk-otel-collector service failed to uninstall" }
}

$sb = [scriptblock]::create("$Target")
Invoke-Command -ScriptBlock $sb
