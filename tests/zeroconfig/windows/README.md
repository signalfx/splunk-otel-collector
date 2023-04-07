# Tests for Splunk Zero Configuration of .NET applications hosted on Windows IIS

Tests in this directory validate that automatic instrumentation of
.NET and .NET Framework applications hosted on Windows IIS are working under
a deployment following the Zero Configuration procedures.

## Requirements to run the tests locally

The GitHub job `windows-zeroconfig-e2e-test` in the [`build-and-test`](../../../.github/workflows/build-and-test.yml)
has the actual requirements to run the tests locally, overall one needs:

- Set up the [`./testdata/docker-setup`](./testdata/docker-setup) folder by adding:
  - The Splunk OpenTelemetry Collector MSI.
  - The [PowerShell install script](../../../internal/buildscripts/packaging/installer/install.ps1).
- Windows OS:
  - .NET Framework
  - .NET SDK
  - NuGet command-line tool
  - Docker configured to run Windows containers
