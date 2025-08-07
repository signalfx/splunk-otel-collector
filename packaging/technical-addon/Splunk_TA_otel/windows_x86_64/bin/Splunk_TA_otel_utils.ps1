$otelProcessName=$args[0]

function otelLogWrite
{
   Param ([string]$logstring)
   $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff K"

   # Write-Warning is much less verbose than Write-Error and the other Write-* may end up
   # not being captured by the parent process (expected to be the cmd that launched the script)
   Write-Warning "$timestamp Splunk_TA_otel_utils.ps1: $logstring"
}

function getParentProcessId($id)
{
	return (Get-WmiObject Win32_Process | Where-Object ProcessId -eq $id).ParentProcessId
}

function getParentAndGrandParent($id)
{
	try {
		$parentId = getParentProcessId($id)
		otelLogWrite "INFO Parent process id: $parentId"
		$grandParentId = getParentProcessId($parentId)
		otelLogWrite "INFO GrandParent process id: $grandParentId"		
	}
	catch {
		# Use Write-Error directly here to get detailed information.
		Write-Error "ERROR getParentAndGrandParent: $_"
	}

	return $parentId, $grandParentId
}

function getOtelProcessPid($processName, $parentId)
{
	otelLogWrite "INFO Getting PID for $processName with parent PID $parentId ..."
	$procPid = (Get-WmiObject Win32_Process | Where-Object ParentProcessId -eq $parentId | Where-Object ProcessName -eq $processName).ProcessId
	otelLogWrite "INFO Collector process id: $procPid"
	$otelProcess = Get-Process -Id $procPid -ErrorAction Ignore
	if($otelProcess)
	{
		return $procPid
	}
	else
	{
		return -1
	}
}

function forceStopProcess($processId)
{
	otelLogWrite "INFO Stopping $processId if stil running"
	# Use a pipeline to avoid an error message if the process already finished
	Get-WmiObject Win32_Process | Where-Object ProcessId -eq $processId | ForEach-Object { taskkill.exe /f /pid $_.ProcessId }
}

otelLogWrite "INFO Splunk_TA_otelutils.ps1 started"
try {
	otelLogWrite "INFO sleeping ..."
	start-sleep -s 3
	otelLogWrite "INFO getting parent and grand parent PIDs"
	$parentId, $grandParentId = getParentAndGrandParent $PID
	otelLogWrite "INFO got parent $parentId and grand parent $grandParentId PIDs"
	$otelPid = getOtelProcessPid $otelProcessName $parentId
	if ($otelPid -eq -1) {
		otelLogWrite "ERROR Otel agent not running"
	} else {
		otelLogWrite "INFO Otel agent running"
		$parentJob      = Start-Job { Wait-Process -Id $using:parentId } -Name "cmd job"
		$grandParentJob = Start-Job { Wait-Process -Id $using:grandParentId } -Name "splunkd job"
		$otelJob        = Start-Job { Wait-Process -Id $using:otelPid } -Name "Otel agent job"
		otelLogWrite "INFO waiting on termination of any one of splunkd, cmd, and Otel agent"
		$finishedJob = Wait-Job -Any -Job $parentJob,$grandParentJob,$otelJob
		otelLogWrite "INFO at least one of the jobs finished:`n`t'$($finishedJob.Name)': finished with $($finishedJob.JobStateInfo)"
		forceStopProcess $otelPid
		forceStopProcess $parentId
	}	
}
catch {
	# Use Write-Error directly here to get detailed information.
	Write-Error "ERROR $_"
}
finally {
	otelLogWrite "INFO finished"
}
