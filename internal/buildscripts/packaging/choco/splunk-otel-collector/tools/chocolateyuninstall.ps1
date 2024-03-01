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
        Write-Information "Multiple entries found for $softwareName. This can happen when versions prior to 0.95.0 were installed."
    }

    Write-Information "Uninstalling $softwareName ..."

    while ($key.Count -ge 1) {
        Write-Debug "Uninstalling Chocolatey Package $key"
        $curr = $key[0]
        $packageArgs['file'] = "$($curr.UninstallString)" #NOTE: You may need to split this if it contains spaces, see below
        if ($packageArgs['fileType'] -eq 'MSI') {
            $packageArgs['silentArgs'] = "$($curr.PSChildName) $($packageArgs['silentArgs'])"
            $packageArgs['file'] = ''
        }
        Uninstall-ChocolateyPackage @packageArgs

        [array]$key = Get-UninstallRegistryKey -SoftwareName $softwareName
    }

    Write-Information "$softwareName was uninstalled."
}
