$otelProcessName=$args[0]
$splunkTAOtelLogFileName=$args[1]
$splunkTAOtelLogDir=Split-Path -Path $splunkTAOtelLogFileName
$splunkTAPSLogFile=Join-Path -Path $splunkTAOtelLogDir -ChildPath Splunk_TA_otelutils.log

function otelLogWrite
{
   Param ([string]$logstring)
   $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff K"
   Add-content $splunkTAPSLogFile -value "$timestamp $logstring"
}

function isOtelProcessRunning($processName)
{
	$parent = (Get-WmiObject win32_process | ? processid -eq $PID).parentprocessid
	$Global:parentPid = $parent
	otelLogWrite "INFO Parent process id: $parent"
	$grandParent = (Get-WmiObject win32_process | ? processid -eq $parentPid).parentprocessid
	$Global:grandParentPid = $grandParent
	otelLogWrite "INFO GrandParent process id: $Global:grandParentPid"
	$child = (Get-WmiObject win32_process | ? parentprocessid -eq $parent | ? processname -eq $processName).processid
	$Global:otelPid = $child
	otelLogWrite "INFO Collector process id: $otelPid"
	$otelProcess = Get-Process -Id $child
	if($otelProcess)
	{
		return $true
	}
	else
	{
		return $false
	}
}

function waitForExit($parent) 
{
	Wait-Process -Id $parent
	otelLogWrite "INFO Parent process exited"
}

function forceStopOtelProcess($processId)
{
	Stop-Process -Id $processId -Force
	Start-Sleep -Seconds 1
}

function CheckOtelProcessStop($processId)
{
	$processInfo = Get-WmiObject win32_process | ? processid -eq $processId
	if($processInfo) {
		otelLogWrite "ERROR Otel agent didn't stop"
	} else {
		otelLogWrite "INFO Otel agent stopped"
	}
}

otelLogWrite "INFO Splunk_TA_otelutils.ps1 started"
start-sleep -s 3
if (isOtelProcessRunning($otelProcessName)) {
	otelLogWrite "INFO Otel agent running"
	waitForExit($grandParentPid)
	forceStopOtelProcess($otelPid)
	CheckOtelProcessStop($otelPid)
} else {
	otelLogWrite "ERROR Otel agent not running"
}
