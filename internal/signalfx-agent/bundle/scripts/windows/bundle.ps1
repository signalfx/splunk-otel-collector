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

Add-Type -AssemblyName System.IO.Compression.FileSystem
$scriptDir = split-path -parent $MyInvocation.MyCommand.Definition
$repoDir = "$scriptDir\..\..\..\..\.."
. $scriptDir\common.ps1

$BUILD_DIR="$repoDir\build"
$PYTHON_VERSION="3.11.3"
$PIP_VERSION="21.0.1"
$NUGET_URL="https://aka.ms/nugetclidl"
$NUGET_EXE="nuget.exe"

# download collectd from github.com/signalfx/collectd
function download_collectd([string]$collectdCommit, [string]$outputDir="$BUILD_DIR\collectd") {
    mkdir $outputDir -ErrorAction Ignore
    download_file -url "https://github.com/signalfx/collectd/archive/$collectdCommit.zip" -outputDir $outputDir -fileName "collectd.zip"
}

function get_collectd_plugins ([string]$buildDir=$BUILD_DIR) {
    mkdir "$buildDir\collectd-python" -ErrorAction Ignore
    $collectdPlugins = Resolve-Path "$buildDir\collectd-python"
    $requirements = Resolve-Path "$scriptDir\..\requirements.txt"
    $script = Resolve-Path "$scriptDir\..\get-collectd-plugins.py"
    $python = Resolve-Path "$buildDir\python\python.exe"
    & $python -m pip install -qq -r $requirements
    if ($lastexitcode -ne 0){ throw }
    & $python $script $collectdPlugins
    if ($lastexitcode -ne 0){ throw }
    & $python -m pip list
    & $python -m pip uninstall pip -y
    if ($lastexitcode -ne 0){ throw }
}

function download_nuget([string]$url=$NUGET_URL, [string]$outputDir=$BUILD_DIR) {
    Remove-Item -Force "$outputDir\$NUGET_EXE" -ErrorAction Ignore
    download_file -url $url -outputDir $outputDir -fileName $NUGET_EXE
}

function install_python([string]$buildDir=$BUILD_DIR, [string]$pythonVersion=$PYTHON_VERSION, [string]$pipVersion=$PIP_VERSION) {
    $nugetPath = Resolve-Path -Path "$buildDir\$NUGET_EXE"
    $installPath = "$buildDir\python.$pythonVersion"
    $targetPath = "$buildDir\python"

    Remove-Item -Recurse -Force $installPath -ErrorAction Ignore
    Remove-Item -Recurse -Force $targetPath -ErrorAction Ignore

    & $nugetPath locals all -clear

    if (((& $nugetPath sources list 2> $null) | Select-String "nuget.org") -Eq $null) {
        & $nugetPath sources add -name "nuget.org" -source "https://api.nuget.org/v3/index.json"
    }

    & $nugetPath install python -Version $pythonVersion -OutputDirectory $buildDir
    mv "$installPath\tools" $targetPath

    Remove-Item -Recurse -Force $installPath

    & $targetPath\python.exe -m pip install pip==$pipVersion --no-warn-script-location
    & $targetPath\python.exe -m ensurepip
}

# install sfxpython package from the local directory
function bundle_python_runner($buildDir=".\build") {
    $python = Resolve-Path -Path "$buildDir\python\python.exe"
    $arguments = "-m", "pip", "install", "-qq", "$scriptDir\..\..\python", "--upgrade"
    & $python $arguments
    if ($lastexitcode -ne 0){ throw }
}
