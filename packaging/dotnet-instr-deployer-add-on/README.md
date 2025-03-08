# Splunk OpenTelemetry .NET Deployer

The Splunk OpenTelemetry .NET Deployer is a technical add-on to be used with
Splunk Enterprise to facilitate the deployment of the
[Splunk Distribution of OpenTelemetry .NET](https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/get-started.html).
It automates the setup process, ensuring that your .NET
applications are properly configured to send telemetry data to Splunk
Observability Cloud.

It assumes that the Splunk OpenTelemetry Collector is already deployed and
configured to receive telemetry data from your .NET applications.

## Installation

It follows the standard installation process for Splunk apps and add-ons.
Notice that the technical add-on is not a standalone application, it is meant
to be run under `splunkd` control.

__Note:__ the technical add-on is only available for Windows x86_64 and requires
`splunkd` to be running under a user with administrative privileges. If the
`splunkd` process is not running with the necessary rights, the installation
will fail with an error similar to:

```PowerShell
+ Import-Module $modulePath
+ ~~~~~~~~~~~~~~~~~~~~~~~~~
    + CategoryInfo          : PermissionDenied: (Splunk.OTel.DotNet.psm1:String) [Import-Module], ScriptRequiresException
    + FullyQualifiedErrorId : ScriptRequiresElevation,Microsoft.PowerShell.Commands.ImportModuleCommand
```

## Configuration

See the [inputs.conf.spec](./assets/README/inputs.conf.spec) file for the available
configuration options.

## Development

This project was developed on Windows, using `MINGW64` tools and the `bash`
provided by Git for Windows. Check the [Makefile](./Makefile) for the available
targets - use the `build-pack` target to generate the technical add-on under
`./out/distribution`.

### Sources

#### `./assets/`

Mimics the structure of the technical add-on, notice the
[`.gitignore`](./.gitignore) file to
identify which assets are built and which are not.

It can be copied directly to a Splunk Enterprise instance to install the
technical add-on. __Note:__ the name of the directory must be the same as the
name of the technical add-on.

#### `./cmd/splunk_otel_dotnet_deployer/`

Contains the source code for the deployer. It is a simple Go application that
reads the input XML from the `stdin` and calls a wrapper PowerShell script to
control the installation process.

#### `./internal/`

Contains the internal packages used by the deployer that in principle could
be used by other technical add-ons. Currently only contains a type to read the
input XML.

### Manual run of the deployer

To manually run the deployer locally, use the following command:

```PowerShell
> cd .\assets\windows_x86_64\bin

> Get-Content ..\..\..\internal\modularinput\testdata\install.inputs.xml | .\splunk_otel_dotnet_deployer.exe
```

The application will generate output on the console. If you also want the log
file you need to define the `SPLUNK_HOME` environment variable, pointing to an
existing directory. The log file will be created under
`$SPLUNK_HOME/var/log/splunk/splunk_otel_dotnet_deployer.log`. You need to
create the directory structure, including 'var/log/splunk' if it does not exist.
