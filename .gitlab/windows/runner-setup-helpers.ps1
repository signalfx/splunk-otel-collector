<#
.SYNOPSIS
    Downloads a Chocolatey package (.nupkg) file.

.DESCRIPTION
    This function downloads a Chocolatey package file from a specified source. 
    It is typically used during GitLab runner setup to obtain required packages 
    for the Windows build environment.

.NOTES
    This function is part of the GitLab Windows runner setup helpers and is 
    intended for use in CI/CD pipeline configurations.
#>
function Get-ChocolateyNupkg {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$PackageName,

        [Parameter(Mandatory=$true)]
        [string]$Version,

        [Parameter(Mandatory=$true)]
        [string]$DestinationDirectory
    )

    # Ensure destination directory exists
    if (-not (Test-Path -Path $DestinationDirectory)) {
        New-Item -ItemType Directory -Path $DestinationDirectory -Force | Out-Null
        Write-Host "Created directory: $DestinationDirectory"
    }

    # Construct the download URL and destination path
    $baseUrl = "https://community.chocolatey.org/api/v2/package"
    $downloadUrl = "$baseUrl/$PackageName/$Version"
    $destinationFile = Join-Path -Path $DestinationDirectory -ChildPath "$PackageName.$Version.nupkg"

    Write-Host "Downloading $PackageName version $Version..."
    Write-Host "From: $downloadUrl"
    Write-Host "To: $destinationFile"

    try {
        $ProgressPreferenceBackup = $ProgressPreference
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $destinationFile -UseBasicParsing
        Write-Host "Successfully downloaded $PackageName.$Version.nupkg"
        return $destinationFile
    }
    catch {
        Write-Error "Failed to download $PackageName.$Version.nupkg: $_"
        throw
    }
    finally {
        $ProgressPreference = $ProgressPreferenceBackup
    }
}

<#
.SYNOPSIS
    Verifies the SHA256 hash of a file. Throws an error if the hash does not match.

.DESCRIPTION
    This function calculates the SHA256 hash of a specified file and compares it 
    against an expected hash value. If the hashes do not match, the function throws 
    an error. This is used to ensure the integrity of downloaded packages.

.NOTES
    This function is part of the GitLab Windows runner setup helpers and is 
    intended for use in CI/CD pipeline configurations.
#>
function Confirm-Sha256 {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$FilePath,

        [Parameter(Mandatory=$true)]
        [string]$ExpectedHash
    )

    # Verify the file exists
    if (-not (Test-Path -Path $FilePath)) {
        throw "File not found: $FilePath"
    }

    Write-Host "Verifying SHA256 hash for: $FilePath"

    try {
        # Calculate the actual hash
        $actualHash = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash

        # Normalize both hashes to lowercase for comparison
        $actualHash = $actualHash.ToLower()
        $ExpectedHash = $ExpectedHash.ToLower()

        Write-Host "Expected hash: $ExpectedHash"
        Write-Host "Actual hash:   $actualHash"

        # Compare the hashes
        if ($actualHash -ne $ExpectedHash) {
            throw "SHA256 hash mismatch for file: $FilePath`nExpected: $ExpectedHash`nActual: $actualHash"
        }

        Write-Host "SHA256 hash verification successful"
        return $true
    }
    catch {
        Write-Error "Failed to verify SHA256 hash: $_"
        throw
    }
}

<#
.SYNOPSIS
    Installs a hard-coded version of Chocolatey.

.DESCRIPTION
    This function installs a hard-coded version of Chocolatey, by downloading the 
    corresponding .nupkg file, verifying its SHA256 hash, and executing the
    installation script. It also configures Chocolatey to use a local source
    pointing to the directory where the .nupkg files are stored. The location of
    this directory is made available via a global variable for use by other functions,
    This variable is named $ChocolateyNupkgDirectory.

.NOTES
    This function is part of the GitLab Windows runner setup helpers and is 
    intended for use in CI/CD pipeline configurations.
#>
function Install-Chocolatey {
    $installScriptPath = Join-Path -Path $PSScriptRoot -ChildPath "install-chocolatey.ps1"

    # Validate the install script signature and add signer to Trusted Publishers store
    $sig = Get-AuthenticodeSignature $installScriptPath
    if ($sig.Status -eq "Valid") {
        $cert = $sig.SignerCertificate

        # Add to Trusted Publishers store
        $store = New-Object System.Security.Cryptography.X509Certificates.X509Store("TrustedPublisher", "LocalMachine")
        $store.Open("ReadWrite")
        $store.Add($cert)
        $store.Close()

        Write-Host "Added trusted publisher: $($cert.Subject)"
        Write-Host "Thumbprint: $($cert.Thumbprint)"
    } else {
        Write-Error "Script signature is not valid: $($sig.Status)"
        throw
    }

    $chocolateyVersion = "2.6.0"
    $knownHash = "f13a2af9cd4ec2c9b58d81861bc95ad7151e3a871d8f758dffa72a996a3792d8"

    $nupkgDirectory = Join-Path -Path $Env:TEMP -ChildPath "nupkg"
    $chocoNupkgPath = Get-ChocolateyNupkg -PackageName "chocolatey" -Version $chocolateyVersion -DestinationDirectory $nupkgDirectory

    Confirm-Sha256 -FilePath $chocoNupkgPath -ExpectedHash $knownHash

    Write-Host "Installing Chocolatey version $chocolateyVersion from $chocoNupkgPath"

    $originalExecutionPolicy = Get-ExecutionPolicy 
    Set-ExecutionPolicy AllSigned -Scope Process -Force

    try {
        & $installScriptPath -ChocolateyDownloadUrl $chocoNupkgPath
        Write-Host "Chocolatey installation completed successfully"
    }
    catch {
        Write-Error "Failed to install Chocolatey: $_"
        throw
    }
    finally {
        Set-ExecutionPolicy $originalExecutionPolicy -Scope Process -Force
    }

    # Setup Chocolatey sources only to $nupkgDirectory
    choco source remove -n=chocolatey
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to remove default Chocolatey source 'chocolatey'."
    }
    choco source add -n=local-nupkg -s="$nupkgDirectory"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to add local Chocolatey source '$nupkgDirectory'."
    }
    Write-Host "Configured Chocolatey to use local source: $nupkgDirectory"
    Write-Host "Current Chocolatey sources:"
    choco source list

    # Make the local source location available to the environment running the script
    New-Variable -Name "ChocolateyNupkgDirectory" -Value $nupkgDirectory -Scope Global -Force -Option AllScope,ReadOnly
}

<#
.SYNOPSIS
    Confirms that Chocolatey is installed by checking for the presence of the
    $ChocolateyNupkgDirectory variable.
#>
function Confirm-ChocolateyInstallation {
    if (-not $ChocolateyNupkgDirectory) {
        throw "ChocolateyNupkgDirectory variable is not set. Chocolatey may not be installed or was incorrectly configured."
    }
}

<#
.SYNOPSIS
    Installs a specified Chocolatey package after downloading and verifying its .nupkg file.
#>
function Install-ChocolateyPackage {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$PackageName,

        [Parameter(Mandatory=$true)]
        [string]$Version,

        [Parameter(Mandatory=$true)]
        [string]$ExpectedHash
    )

    Confirm-ChocolateyInstallation

    $nupkgPath = Get-ChocolateyNupkg -PackageName $PackageName -Version $Version -DestinationDirectory $ChocolateyNupkgDirectory
    Confirm-Sha256 -FilePath $nupkgPath -ExpectedHash $ExpectedHash

    Write-Host "Installing $PackageName version $Version via Chocolatey..."
    choco install $PackageName --version $Version --force --yes --source="local-nupkg"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to install $PackageName version $Version via Chocolatey."
    }
}

<#
.SYNOPSIS
    Installs a hard-coded version of Docker CLI via Chocolatey.
#>
function Install-DockerCli {
    Install-ChocolateyPackage `
        -PackageName "docker-cli" `
        -Version "29.1.3" `
        -ExpectedHash "eebdce6e042106b1e6bbcaabd7681417eeee818007fd96a280a3b6d45c17e895"
}

<#
.SYNOPSIS
    Installs a hard-coded version of WiX Toolset via Chocolatey.
#>
function Install-WiXToolset {
    # WixToolset depends on .NET Framework 3.5. Download the package, verify its hash,
    # but let the WixToolset installer handle the actual installation of .NET 3.5.
    $dotnetNupkgPath = Get-ChocolateyNupkg `
        -PackageName "dotnet3.5" `
        -Version "3.5.20241212" `
        -DestinationDirectory $ChocolateyNupkgDirectory
    Confirm-Sha256 `
        -FilePath $dotnetNupkgPath `
        -ExpectedHash "648d30a966fcd9d3f88845b87796fd9492568d3276fee1c123003dbe41b27dca"

    Install-ChocolateyPackage `
        -PackageName "wixtoolset" `
        -Version "3.14.0" `
        -ExpectedHash "d0a5dbfb079fa372c4f35514cb3a80e68b4de5ec31c28af071ba79b6c20baec3"
}

<#
.SYNOPSIS
    Installs a hard-coded version of Git via Chocolatey.
#>
function Install-Git {
    # Installing Git via Chocolatey requires the chocolatey-core.extension package.
    #  Download the package, verify its hash,, but let the git.install package handle
    # the actual installation of the core extension.
    $coreExtNupkgPath = Get-ChocolateyNupkg `
        -PackageName "chocolatey-core.extension" `
        -Version "1.3.3" `
        -DestinationDirectory $ChocolateyNupkgDirectory
    Confirm-Sha256 `
        -FilePath $coreExtNupkgPath `
        -ExpectedHash "ed7a281f45a61150df0e1414651bca8501004c56deb43439c075f1ba58aff70a"

    Install-ChocolateyPackage `
        -PackageName "git.install" `
        -Version "2.52.0" `
        -ExpectedHash "acb81afb45e64dcde2c89c5c04228f9d4fb140e93059224b21687bbd8159a82c"
}
