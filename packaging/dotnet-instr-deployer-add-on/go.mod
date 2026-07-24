module github.com/splunk/splunk_otel_dotnet_deployer

go 1.26.5

require github.com/signalfx/splunk-otel-collector v0.0.0

require (
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/lufia/plan9stats v0.0.0-20260330125221-c963978e514e // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v4 v4.26.6 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/signalfx/splunk-otel-collector => ../..
