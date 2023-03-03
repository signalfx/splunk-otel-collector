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
.PARAMETER Target
    Build target to run (bundle)
#>
param(
    [Parameter(Mandatory=$true, Position=1)][string]$Target,
    [Parameter(Mandatory=$false, ValueFromRemainingArguments=$true)]$Remaining
)

Set-PSDebug -Trace 1
$BUNDLE_NAME = "agent-bundle_windows_amd64.zip"
$BUNDLE_DIR = "agent-bundle"
$COLLECTD_VERSION = "5.8.0-sfx0"
$COLLECTD_COMMIT = "4d3327b14cf4359029613baf4f90c4952702105e"
$ErrorActionPreference = "Stop"

$scriptDir = split-path -parent $MyInvocation.MyCommand.Definition
$repoDir = "$scriptDir\..\..\.."

. "$scriptDir\common.ps1"
. "$scriptDir\bundle.ps1"

# make the build bundle
function bundle (
        [string]$buildDir="$repoDir\build",
        [string]$outputDir="$repoDir\dist",
        [bool]$DOWNLOAD_PYTHON=$false,
        [bool]$DOWNLOAD_COLLECTD=$false,
        [bool]$DOWNLOAD_COLLECTD_PLUGINS=$false) {
    mkdir "$buildDir\$BUNDLE_DIR" -ErrorAction Ignore
    Remove-Item -Recurse -Force "$buildDir\$BUNDLE_DIR\*" -ErrorAction Ignore

    if ($DOWNLOAD_PYTHON -Or !(Test-Path -Path "$buildDir\python")) {
        Remove-Item -Recurse -Force "$buildDir\python" -ErrorAction Ignore
        download_nuget -outputDir $buildDir
        install_python -buildDir $buildDir
    }

    if ($DOWNLOAD_COLLECTD_PLUGINS -Or !(Test-Path -Path "$buildDir\collectd-python")) {
        Remove-Item -Recurse -Force "$buildDir\collectd-python" -ErrorAction Ignore
        bundle_python_runner -buildDir "$buildDir"
        get_collectd_plugins -buildDir "$buildDir"
    }

    if ($DOWNLOAD_COLLECTD -Or !(Test-Path -Path "$buildDir\collectd")) {
        Remove-Item -Recurse -Force "$buildDir\collectd" -ErrorAction Ignore
        mkdir "$buildDir\collectd" -ErrorAction Ignore
        download_collectd -collectdCommit $COLLECTD_COMMIT -outputDir "$buildDir"
        unzip_file -zipFile "$buildDir\collectd.zip" -outputDir "$buildDir\collectd"
    }

    # copy python into agent-bundle directory
    Copy-Item -Path "$buildDir\python" -Destination "$buildDir\$BUNDLE_DIR\python" -recurse -Force
    # copy Python plugins into agent-bundle directory
    Copy-Item -Path "$buildDir\collectd-python" -Destination "$buildDir\$BUNDLE_DIR\collectd-python" -recurse -Force
    # copy types.db file into agent-bundle directory
    Copy-Item -Path "$buildDir\collectd\collectd-$COLLECTD_COMMIT\src\types.db" "$buildDir\$BUNDLE_DIR\types.db" -Force

    # remove unnecessary files and directories
    Get-ChildItem -recurse -path "$buildDir\$BUNDLE_DIR\*" -include __pycache__ | Remove-Item -force -Recurse
    Get-ChildItem -recurse -path "$buildDir\$BUNDLE_DIR\*" -include *.key,*.pem | Where-Object { $_.Directory -match 'test' } | Remove-Item -force
    Get-ChildItem -recurse -path "$buildDir\$BUNDLE_DIR\*" -include *.pyc,*.pyo,*.whl | Remove-Item -force

    # clean up empty directories
    remove_empty_directories -buildDir "$buildDir\$BUNDLE_DIR"

    mkdir "$outputDir" -ErrorAction Ignore
    Remove-Item -Force "$outputDir\$BUNDLE_NAME" -ErrorAction Ignore
    zip_file -src "$buildDir\$BUNDLE_DIR" -dest "$outputDir\$BUNDLE_NAME"
}

if ($REMAINING.length -gt 0) {
    $sb = [scriptblock]::create("$Target $REMAINING")
    Invoke-Command -ScriptBlock $sb
} else {
    &$Target
}
