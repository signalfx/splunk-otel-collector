ARG BASE_IMAGE=mcr.microsoft.com/windows/servercore:ltsc2022

FROM ${BASE_IMAGE}

SHELL ["powershell", "-NoProfile", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

WORKDIR "C:\\otelcol"
COPY otelcol.exe .\\otelcol.exe
COPY otel-collector-config.yaml .\\otel-collector-config.yaml

WORKDIR "C:\\"
COPY start-example.ps1 .\\start-example.ps1

ENTRYPOINT ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", "C:\\start-example.ps1"]
