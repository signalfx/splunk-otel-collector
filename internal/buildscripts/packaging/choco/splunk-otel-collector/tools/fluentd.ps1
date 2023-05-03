$fluentd_msi_name = "td-agent-4.3.2-x64.msi"
$fluentd_dl_url = "https://packages.treasuredata.com/4/windows/$fluentd_msi_name"

try {
    Resolve-Path $env:TEMP 2>&1>$null
    $tempdir = "${env:TEMP}\Fluentd"
} catch {
    $tempdir = "\tmp\Fluentd"
}

create_temp_dir -tempdir $tempdir
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

install_msi -path "$fluentd_msi_path"

# remove the temporary directory
Remove-Item -Recurse -Force "$tempdir"
