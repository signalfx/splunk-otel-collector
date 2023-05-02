# SignalFx Smart Agent Bundle

The SignalFx Smart Agent Bundle includes the python and java runtimes,
collectd, collectd plugins, and other necessary files required for the
Smart Agent [Extension](
https://github.com/signalfx/splunk-otel-collector/tree/main/pkg/extension/smartagentextension)
and [Receiver](
https://github.com/signalfx/splunk-otel-collector/tree/main/pkg/receiver/smartagentreceiver).

This bundle is included by default in the `splunk-otel-collector`
Linux amd64/x86_64/arm64 and Windows amd64 packages and docker images, and is
installed to:

- Linux deb/rpm packages and docker images:
  - `/usr/lib/splunk-otel-collector/agent-bundle`
- Linux tar.gz package:
  - `<extracted directory>/splunk-otel-collector/agent-bundle`
- Windows msi/nupkg/choco packages and docker images:
  - `C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle`

> **Note:** Support for the Linux arm64 bundle is currently **experimental**.

## Manual Installation

The Linux (`agent-bundle_<VERSION>_linux_amd64.tar.gz`) and Windows
(`agent-bundle_<VERSION>_windows_amd64.zip`) bundles can be manually downloaded
and installed from [GitHub Releases](
https://github.com/signalfx/splunk-otel-collector/releases).

### Linux

1. After downloading the Linux bundle, run the following commands to
   extract the bundle and ensure that the binaries in the bundle have the right
   loader set on them since your host's loader may not be compatible.
   ```sh
   $ tar -xzf agent-bundle_<VERSION>_linux_amd64.tar.gz
   $ cd agent-bundle
   $ bin/patch-interpreter $(pwd)
   ```
2. Add/Update `bundleDir` and `collectd/configDir` of the `smartagent`
   extension in your collector configuration file to the absolute path of the
   extracted `agent-bundle` directory before running the collector.  For
   example:
   ```yaml
   extensions:
     smartagent:
       bundleDir: "/path/to/extracted/directory/agent-bundle"
       collectd:
         configDir: "/path/to/extracted/directory/agent-bundle/run/collectd"
   ```

### Windows

1. After downloading the Windows bundle, extract the zip file with the tool of
   your choice or run the following Powershell command:
   ```sh
   PS> Expand-Archive -Path "C:\path\to\download\folder\agent-bundle_<VERSION>_windows_amd64.zip" -DestinationPath "C:\path\to\extracted\folder"
   ```
2. Add/Update `bundleDir` and `collectd/configDir` of the `smartagent`
   extension in your collector configuration file to the absolute path of the
   extracted `agent-bundle` directory before running the collector.  For
   example:
   ```yaml
   extensions:
     smartagent:
       bundleDir: "C:\\path\\to\\extracted\\folder\\agent-bundle"
       collectd:
         configDir: "C:\\path\\to\\extracted\\folder\\agent-bundle\\run\\collectd"
   ```

## Development

### Linux

Run the following commands to build the bundle for Linux (requires `git`,
`docker`, `buildkit`, and `make`):
```sh
$ git clone https://github.com/signalfx/splunk-otel-collector
$ cd splunk-otel-collector
$ make -C internal/signalfx-agent/bundle agent-bundle-linux ARCH=<amd64|arm64>
```

The bundle will be saved to `dist/agent-bundle_linux_<amd64|arm64>.tar.gz`.

> **Note:** Support for the Linux arm64 bundle is currently **experimental**.

### Windows

Run the following Powershell commands to build the bundle for Windows (requires
`git`, .NET Framework 4.5 or newer, and Powershell 5.0 or newer)
```sh
PS> git clone https://github.com/signalfx/splunk-otel-collector
PS> cd splunk-otel-collector
PS> ./internal/signalfx-agent/bundle/scripts/windows/make.ps1 bundle
```
The bundle will be saved to `dist/agent-bundl_windows_amd64.zip`.
