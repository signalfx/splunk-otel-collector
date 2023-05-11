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

Set-PSDebug -Trace 1
$MY_SCRIPT_DIR = $scriptDir = split-path -parent $MyInvocation.MyCommand.Definition

# https://blog.jourdant.me/post/3-ways-to-download-files-with-powershell
function download_file([string]$url, [string]$outputDir, [string]$fileName) {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    (New-Object System.Net.WebClient).DownloadFile($url, "$outputDir\$fileName")
}

function unzip_file($zipFile, $outputDir){
    # this requires .net 4.5 and above
    Add-Type -assembly "system.io.compression.filesystem"
    [System.IO.Compression.ZipFile]::ExtractToDirectory($zipFile, $outputDir)
}

function zip_file($src, $dest) {
    # this requires .net 4.5 and above
    Add-Type -assembly "system.io.compression.filesystem"
    $SRC = Resolve-Path -Path $src
    [System.AppContext]::SetSwitch('Switch.System.IO.Compression.ZipFile.UseBackslash', $false)
    [System.IO.Compression.ZipFile]::CreateFromDirectory($SRC, "$dest", 1, $true)
}

function remove_empty_directories ($buildDir) {
    Set-PSDebug -Trace 0
    do {
        $dirs = gci $buildDir -directory -recurse | Where { (gci $_.fullName -Force).count -eq 0 } | select -expandproperty FullName
        $dirs | Foreach-Object { Remove-Item $_ }
    } while ($dirs.count -gt 0)
    Set-PSDebug -Trace 1
}

function replace_text([string]$filepath, [string]$find, [string]$replacement) {
    (Get-Content $filepath).replace($find, $replacement) | Set-Content $filepath
}
