ARG BASE_IMAGE=mcr.microsoft.com/windows/servercore:ltsc2022

FROM ${BASE_IMAGE}

SHELL ["powershell", "-NoProfile", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

WORKDIR "C:\\otelcol"
COPY otelcol.exe .\\otelcol.exe
COPY otel-collector-config.yaml .\\otel-collector-config.yaml

WORKDIR "C:\\backup-job"
COPY backup-job\\ .\\

RUN New-EventLog -LogName Application -Source BackupJob

WORKDIR "C:\\"
COPY start-example.ps1 .\\start-example.ps1

ENV INFLUXDB_ENDPOINT="http://127.0.0.1:8086/api/v2/write?org=example-org&bucket=example-bucket&precision=s"
ENV OTLP_ENDPOINT="http://127.0.0.1:4318/v1/metrics"

ENTRYPOINT ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", "C:\\start-example.ps1"]
