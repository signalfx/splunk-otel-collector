$otelProcessName=$args[0]

$scriptName = $MyInvocation.MyCommand.Name

function Write-Log
{
	Param ([string]$msg)
	$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff K"

	# Write-Warning is much less verbose than Write-Error and the other Write-* may end up
	# not being captured by the parent process (expected to be the cmd that launched the script)
	Write-Warning "${scriptName}: $timestamp $msg"
}

function Get-Win32_ProcessByNameAndParentId
{
	Param (
		[string]$processName,
		[Int32]$parentId,
		[TimeSpan]$timeout,
		[TimeSpan]$sleepInterval)

	$process = $null
	$startTime = Get-Date

	do {
		$process = Get-WmiObject Win32_Process | Where-Object ParentProcessId -eq $parentId | Where-Object ProcessName -eq $processName
		if ($process) {
			break
		}

		Start-Sleep -Milliseconds ($sleepInterval.TotalMilliseconds)
	} while (((Get-Date) - $startTime) -lt $timeout)

	return $process
}

function Stop-ProcessById
{
	Param ([Int32]$processId)

	Write-Log "INFO Stopping PID $processId if still running"
	# Use a pipeline to avoid an error message if the process already finished
	Get-WmiObject Win32_Process | Where-Object ProcessId -eq $processId | ForEach-Object { taskkill.exe /f /pid $_.ProcessId }
}

function Get-Win32_Process {
	Param([Int32]$processId)
	return Get-WmiObject Win32_Process -Filter "ProcessId = $processId"
}

Write-Log "INFO started"
try {
	$currentProcess     = Get-Win32_Process $PID # $PID is predefined as the PID of the current process
	$parentProcess      = Get-Win32_Process $currentProcess.ParentProcessId
	$grandParentProcess = Get-Win32_Process $parentProcess.ParentProcessId

	# splunkd launched a cmd that launched the collector, but didn't wait on it and then launched this script
	# get the expected process for the OTel collector launched via splunkd
	$otelProcess = Get-Win32_ProcessByNameAndParentId `
		$otelProcessName `
		$parentProcess.ProcessId `
		-timeout ([TimeSpan]::FromSeconds(10)) `
		-sleepInterval ([TimeSpan]::FromMilliseconds(150))
	if ($null -eq $otelProcess) {
		Write-Log "ERROR Otel agent not running"
		return
	}

	Write-Log "INFO Otel agent running (PID $($otelProcess.ProcessId))"
	$parentJob      = Start-Job { Wait-Process -Id $using:parentProcess.ProcessId }      -Name "cmd job"
	$grandParentJob = Start-Job { Wait-Process -Id $using:grandParentProcess.ProcessId } -Name "splunkd job"
	$otelJob        = Start-Job { Wait-Process -Id $using:otelProcess.ProcessId }        -Name "Otel agent job"

	Write-Log "INFO waiting on termination of splunkd, cmd, or the Otel agent"
	$finishedJob = Wait-Job -Any -Job $parentJob,$grandParentJob,$otelJob

	Write-Log "INFO at least one of the jobs finished:`n`t'$($finishedJob.Name)': finished with $($finishedJob.JobStateInfo)"
	Stop-ProcessById $otelProcess.ProcessId
	Stop-ProcessById $parentProcess.ProcessId
}
catch {
	# Use Write-Error directly here to get detailed information.
	Write-Error "ERROR $_"
}
finally {
	Write-Log "INFO finished"
}
