$otelProcessName=$args[0]
$splunkTAOtelLogFileName=$args[1]
$splunkTAOtelLogDir=Split-Path -Path $splunkTAOtelLogFileName
$splunkTAPSLogFile=Join-Path -Path $splunkTAOtelLogDir -ChildPath Splunk_TA_otelutils.log

function otelLogWrite
{
   Param ([string]$logstring)

   Add-content $splunkTAPSLogFile -value $logstring
}

function isOtelProcessRunning($processName)
{
	$parent = (Get-WmiObject win32_process | ? processid -eq $PID).parentprocessid
	$Global:parentPid = $parent
	$child = (Get-WmiObject win32_process | ? parentprocessid -eq $parent | ? processname -eq $processName).processid
	$Global:otelPid = $child
	$otelProcess = Get-Process -Id $child
	if($otelProcess)
	{
		otelLogWrite "INFO Otel agent running"
		return 1
	}
}

function waitForExit($parent) 
{
	Wait-Process -Id $parent
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

start-sleep -s 3
isOtelProcessRunning($otelProcessName)

waitForExit($parentPid)
forceStopOtelProcess($otelPid)
CheckOtelProcessStop($otelPid)