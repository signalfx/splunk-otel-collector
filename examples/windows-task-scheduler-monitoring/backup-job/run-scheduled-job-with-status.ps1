# Generic Windows Scheduled Task status reporter.
#
# Configure this wrapper as the action of any scheduled task to push a
# success/failure metric for its last run without changing the job's own
# script or executable at all:
#
#   powershell.exe -NoProfile -ExecutionPolicy Bypass -File C:\path\run-scheduled-job-with-status.ps1 -TaskName \YourTask -FilePath C:\path\your-existing-job.ps1
#
# The wrapper runs the target .ps1, .cmd/.bat, or .exe, maps its process exit
# code to success/failure metrics, pushes those metrics to the collector over
# OTLP HTTP, and exits with the same code so Task Scheduler still records the
# original result.

param(
    [Parameter(Mandatory = $true)]
    [string] $TaskName,

    [Parameter(Mandatory = $true)]
    [string] $FilePath,

    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]] $JobArguments = @()
)

$ErrorActionPreference = 'Continue'

function ConvertTo-CommandLineArgument {
    param(
        [AllowEmptyString()]
        [string] $Argument
    )

    if ($Argument.Length -gt 0 -and $Argument -notmatch '[\s"]') {
        return $Argument
    }

    $quoted = '"'
    $backslashes = 0

    foreach ($char in $Argument.ToCharArray()) {
        if ($char -eq '\') {
            $backslashes++
            continue
        }

        if ($char -eq '"') {
            $quoted += ('\' * (($backslashes * 2) + 1))
            $quoted += '"'
            $backslashes = 0
            continue
        }

        if ($backslashes -gt 0) {
            $quoted += ('\' * $backslashes)
            $backslashes = 0
        }
        $quoted += $char
    }

    if ($backslashes -gt 0) {
        $quoted += ('\' * ($backslashes * 2))
    }

    $quoted += '"'
    return $quoted
}

function Join-CommandLineArguments {
    param(
        [string[]] $Arguments
    )

    return (($Arguments | ForEach-Object { ConvertTo-CommandLineArgument $_ }) -join ' ')
}

function Invoke-ScheduledJob {
    param(
        [string] $TargetPath,
        [string[]] $TargetArguments
    )

    $extension = [System.IO.Path]::GetExtension($TargetPath).ToLowerInvariant()
    switch ($extension) {
        '.ps1' {
            $processPath = 'powershell.exe'
            $processArguments = @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $TargetPath) + $TargetArguments
        }
        { $_ -in @('.cmd', '.bat') } {
            $processPath = 'cmd.exe'
            $processArguments = @('/d', '/c', $TargetPath) + $TargetArguments
        }
        default {
            $processPath = $TargetPath
            $processArguments = $TargetArguments
        }
    }

    $process = Start-Process `
        -FilePath $processPath `
        -ArgumentList (Join-CommandLineArguments $processArguments) `
        -NoNewWindow `
        -Wait `
        -PassThru `
        -ErrorAction Stop

    return $process.ExitCode
}

$otlpEndpoint = if ($env:OTLP_ENDPOINT) {
    $env:OTLP_ENDPOINT
} else {
    'http://127.0.0.1:4318/v1/metrics'
}

try {
    if (-not (Test-Path -LiteralPath $FilePath -PathType Leaf)) {
        throw "File not found: $FilePath"
    }
    $exitCode = Invoke-ScheduledJob -TargetPath $FilePath -TargetArguments $JobArguments
} catch {
    Write-Host "run-scheduled-job-with-status: failed to run ${FilePath}: $($_.Exception.Message)"
    $exitCode = 1
}

$successfulRun = 0
$failedRun = 1
if ($exitCode -eq 0) {
    $successfulRun = 1
    $failedRun = 0
}

Write-Host "run-scheduled-job-with-status: $TaskName exit_code=$exitCode"

$nowUnixNano = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() * 1000000
$payload = @{
    resourceMetrics = @(
        @{
            resource = @{
                attributes = @(
                    @{
                        key = 'scheduler.provider'
                        value = @{ stringValue = 'windows_task_scheduler' }
                    },
                    @{
                        key = 'scheduler.task.name'
                        value = @{ stringValue = $TaskName }
                    }
                )
            }
            scopeMetrics = @(
                @{
                    scope = @{ name = 'run-scheduled-job-with-status.ps1' }
                    metrics = @(
                        @{
                            name = 'windows_task_scheduler.job.successful_run'
                            gauge = @{
                                dataPoints = @(
                                    @{
                                        timeUnixNano = "$nowUnixNano"
                                        asInt = $successfulRun
                                    }
                                )
                            }
                        },
                        @{
                            name = 'windows_task_scheduler.job.failed_run'
                            gauge = @{
                                dataPoints = @(
                                    @{
                                        timeUnixNano = "$nowUnixNano"
                                        asInt = $failedRun
                                    }
                                )
                            }
                        }
                    )
                }
            )
        }
    )
} | ConvertTo-Json -Depth 20 -Compress

try {
    $response = Invoke-WebRequest -UseBasicParsing -Method Post -Uri $otlpEndpoint -ContentType 'application/json' -Body $payload
    Write-Host "run-scheduled-job-with-status: pushed metrics for $TaskName to $otlpEndpoint, HTTP $($response.StatusCode)"
} catch {
    Write-Host "run-scheduled-job-with-status: failed to push metrics for $TaskName to ${otlpEndpoint}: $($_.Exception.Message)"
}

exit $exitCode
