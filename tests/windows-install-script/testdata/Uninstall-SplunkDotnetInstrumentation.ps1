$module_name = "Splunk.OTel.DotNet.psm1"
$download = "https://github.com/signalfx/splunk-otel-dotnet/releases/latest/download/$module_name"
$dotnet_autoinstr_path = Join-Path . $module_name
Write-Host "Downloading .NET Instrumentation installer ..."
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $download -OutFile $dotnet_autoinstr_path -UseBasicParsing
} catch {
    $err = $_.Exception.Message
    $message = "
    An error occurred when trying to download .NET Instrumentation installer from $download. This may be due to a network connectivity issue.
    $err
    "
    throw "$message"
}

Import-Module $dotnet_autoinstr_path

Write-Host "Uninstalling Splunk Distribution of OpenTelemetry .NET ..."
Uninstall-OpenTelemetryCore
