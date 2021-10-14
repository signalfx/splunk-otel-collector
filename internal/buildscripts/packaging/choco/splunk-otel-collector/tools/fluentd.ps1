[bool]$SkipFluend = $FALSE

$fluentd_msi_name = "td-agent-4.1.1-x64.msi"
$fluentd_dl_url = "https://packages.treasuredata.com/4/windows/$fluentd_msi_name"
try {
    Resolve-Path $env:SYSTEMDRIVE
    $fluentd_base_dir = "${env:SYSTEMDRIVE}\opt\td-agent"
} catch {
    $fluentd_base_dir = "\opt\td-agent"
}
$fluentd_config_dir = "$fluentd_base_dir\etc\td-agent"
$fluentd_config_path = "$fluentd_config_dir\td-agent.conf"
$fluentd_service_name = "fluentdwinsvc"

try {
    Resolve-Path $env:TEMP
    $tempdir = "${env:TEMP}\Fluentd"
} catch {
    $tempdir = "\tmp\Fluentd"
}

#Skipping installation of fluentd if already installed
if ((service_installed -name "$fluentd_service_name") -OR (Test-Path -Path "$fluentd_base_dir\bin\fluentd")) {
    $SkipFluend = $TRUE
    Write-Host "The $fluentd_service_name service is already installed. Skipping fluentd installation."
}

if (!$SkipFluend) {

    # download a file to a given destination
    function download_file([string]$url, [string]$outputDir, [string]$fileName) {
        try {
            [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
            (New-Object System.Net.WebClient).DownloadFile($url, "$outputDir\$fileName")
        } catch {
            $err = $_.Exception.Message
            $message = "
            An error occurred while downloading $url
            $err
            "
            throw "$message"
        }
    }

    # create the temp directory if it doesn't exist
    function create_temp_dir($tempdir=$tempdir) {
        if ((Test-Path -Path "$tempdir")) {
            Remove-Item -Recurse -Force "$tempdir"
        }
        mkdir "$tempdir" -ErrorAction Ignore
    }

    $tempdir = create_temp_dir -tempdir $tempdir
    $default_fluentd_config = "$installation_path\fluentd\td-agent.conf"
    $default_confd_dir = "$installation_path\fluentd\conf.d"

    # copy the default fluentd config to $fluentd_config_path if it does not already exist
    if (!(Test-Path -Path "$fluentd_config_path") -And (Test-Path -Path "$default_fluentd_config")) {
        $default_fluentd_config = Resolve-Path "$default_fluentd_config"
        Write-Host "Copying $default_fluentd_config to $fluentd_config_path"
        mkdir "$fluentd_config_dir" -ErrorAction Ignore | Out-Null
        Copy-Item "$default_fluentd_config" "$fluentd_config_path"
    }

    # copy the default source configs to $fluentd_config_dir\conf.d if it does not already exist
    if (Test-Path -Path "$default_confd_dir\*.conf") {
        mkdir "$fluentd_config_dir\conf.d" -ErrorAction Ignore | Out-Null
        $confFiles = (Get-Item "$default_confd_dir\*.conf")
        foreach ($confFile in $confFiles) {
            $name = $confFile.Name
            $path = $confFile.FullName
            if (!(Test-Path -Path "$fluentd_config_dir\conf.d\$name")) {
                Write-Host "Copying $path to $fluentd_config_dir\conf.d\$name"
                Copy-Item "$path" "$fluentd_config_dir\conf.d\$name"
            }
        }
    }
    Write-Host "Downloading $fluentd_dl_url..."
    download_file -url "$fluentd_dl_url" -outputDir "$tempdir" -fileName "$fluentd_msi_name"
    $fluentd_msi_path = (Join-Path "$tempdir" "$fluentd_msi_name")

    Write-Host "Installing $fluentd_msi_path ..."
    Start-Process msiexec.exe -Wait -ArgumentList "/qn /norestart /i `"$fluentd_msi_path`""
    Write-Host "- Done"

    stop_service -name "$fluentd_service_name"

    Write-Host "Starting $fluentd_service_name service..."
    start_service -name "$fluentd_service_name" -config_path "$fluentd_config_path"
    Write-Host "- Started"

    # remove the temporary directory
    Remove-Item -Recurse -Force "$tempdir"
}
