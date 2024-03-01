$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

$packageArgs = @{
    packageName    = $env:ChocolateyPackageName
    softwareName   = $env:ChocolateyPackageTitle
    fileType       = 'MSI'
    silentArgs     = "/qn /norestart"
    validExitCodes = @(0)
}

$softwareName = $packageArgs['softwareName']
[array]$key = Get-UninstallRegistryKey -SoftwareName $softwareName

if ($key.Count -eq 0) {
    Write-Warning "$softwareName has already been uninstalled by other means."
} else {
    if ($key.Count -gt 1) {
        Write-Host "Multiple entries found for $softwareName. This can happen when versions prior to 0.95.0 were installed."
    }

    Write-Host "Uninstalling $softwareName ..."

    while ($key.Count -ge 1) {
        $curr = $key[0]
        Write-Host "Uninstalling Chocolatey Package $curr ..."
        $packageArgs['file'] = "$($curr.UninstallString)" #NOTE: You may need to split this if it contains spaces, see below
        if ($packageArgs['fileType'] -eq 'MSI') {
            $packageArgs['silentArgs'] = "$($curr.PSChildName) $($packageArgs['silentArgs'])"
            $packageArgs['file'] = ''
        }
        Uninstall-ChocolateyPackage @packageArgs

        Write-Host "Uninstall complete for Chocolatey Package $curr ..."
        [array]$key = Get-UninstallRegistryKey -SoftwareName $softwareName
        Write-Host "Remaining entries: $($key.Count)"
    }

    Write-Host "$softwareName was uninstalled."
}
