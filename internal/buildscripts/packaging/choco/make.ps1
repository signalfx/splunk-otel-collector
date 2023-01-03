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
.PARAMETER Target
    Build target to run (build_choco)
#>
Param(
    [Parameter(Mandatory=$true, ValueFromRemainingArguments=$true)][string]$Target
)

$scriptDir = split-path -parent $MyInvocation.MyCommand.Definition
$repoDir = "$scriptDir\..\..\..\.."

function build_choco(
        [string]$Version,
        [string]$MSIFile="",
        [string]$BuildDir="$repoDir\dist",
        [string]$ChocoDir="$scriptDir\splunk-otel-collector") {

    echo "Choco package Version to $Version"

    if ($MSIFile -Eq "") {
        $MSIFile = "$BuildDir\splunk-otel-collector-$Version-amd64.msi"
    }

    if (!(Test-Path -Path "$MSIFile")) {
        throw "$MSIFile not found!"
    }

    $msi_hash = (Get-FileHash "$MSIFile" -Algorithm SHA256 | select -ExpandProperty Hash)

    if (Test-Path -Path "$BuildDir\choco") {
        Remove-Item -Recurse -Force "$BuildDir\choco"
    }
    Copy-Item -Recurse "$ChocoDir" "$BuildDir\choco\splunk-otel-collector"

    # update chocolateyinstall.ps1 with MSI name and hash
    $installer_path = "$BuildDir\choco\splunk-otel-collector\tools\chocolateyinstall.ps1"
    if (!(Test-Path -Path "$installer_path")) {
        throw "$installer_path not found!"
    }
    ((Get-Content -Path "$installer_path" -Raw) -Replace "MSI_NAME", ("$MSIFile" | Split-Path -Leaf)) | Set-Content -Path "$installer_path"
    ((Get-Content -Path "$installer_path" -Raw) -Replace "MSI_HASH", "$msi_hash") | Set-Content -Path "$installer_path"

    # update VERIFICATION.txt with MSI name and hash
    $verification_path = "$BuildDir\choco\splunk-otel-collector\tools\VERIFICATION.txt"
    if (!(Test-Path -Path "$verification_path")) {
        throw "$verification_path not found!"
    }
    ((Get-Content -Path "$verification_path" -Raw) -Replace "MSI_NAME", ("$MSIFile" | Split-Path -Leaf)) | Set-Content -Path "$verification_path"
    ((Get-Content -Path "$verification_path" -Raw) -Replace "MSI_HASH", "$msi_hash") | Set-Content -Path "$verification_path"

    # append LICENSE content to LICENSE.txt in choco package
    Get-Content "scriptDir\..\LICENSE" | Add-Content "$BuildDir\choco\splunk-otel-collector\tools\LICENSE.txt"

    Copy-Item "$MSIFile" "$BuildDir\choco\splunk-otel-collector\tools\"

    $dest = "$BuildDir\splunk-otel-collector.$Version.nupkg"

    if (Test-Path -Path "$dest") {
        Remove-Item -Recurse -Force "$dest"
    }

    choco pack --version=$Version --out "$BuildDir" "$BuildDir\choco\splunk-otel-collector\splunk-otel-collector.nuspec"
    if ($lastexitcode -gt 0){ throw }
    if (!(Test-Path -Path "$dest")) {
        throw "$dest not found!"
    }

    echo "built $BuildDir\splunk-otel-collector.$Version.nupkg"
}

$sb = [scriptblock]::create("$Target")
Invoke-Command -ScriptBlock $sb
