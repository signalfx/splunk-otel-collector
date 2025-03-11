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

### Splunk Modular Input Primer

This section is a very brief introduction to the general architecture of a
Splunk Modular Input. In the Splunk documentation there is some overlap with
a Splunk App and Splunk Add-on.

A Splunk Modular Input is shipped as a `.tgz` file with a specific [directory
structure](https://docs.splunk.com/Documentation/Splunk/latest/AdvancedDev/ModInputsBasicExample)
in which the root directory is the name of the modular input.

The directory structure will contain [platform and architecture](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/custominputs/modinputsoverview/#Platform-and-architecture-support)
specific paths for the modular _script_, which can be an binary executable,
Python source file, or a platform native script, e.g.: `cmd` on Windows or
`sh` on Linux.

The modular input script is invoked by the Splunk host process, `splunkd`, and
it is expected to read the input configuration from `stdin` and support the
[three execution modes specified via the command line](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/custominputs/modinputsexamples/#Script-routines).
The modular input must implement the `--scheme` or `--validate-arguments`
modes. At minimum the modular input script should set its exit code to 0
when executed with either parameter.

The _Execution_ mode is when the modular input is doing the actual work and
the Splunk host process is going to send the input configuration as an XML
document to the modular input script's `stdin`. This XML is generated from the
`<modular_input>/default/inputs.conf` file and with overrides at the
`<modular_input>/local/inputs.conf` file. The schema for the modular input
is defined in the `<modular_input>/README/inputs.conf.spec` file and the
general documentation for the `inputs.conf.spec` is available
[here](https://docs.splunk.com/Documentation/Splunk/latest/Admin/Inputsconf).
The actual XML sent by the Splunk host can be obtained by following the
instructions in the [Splunk documentation](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/custominputs/modinputsexamples/#Testing-the-script).

Depending on the definition of the modular input, the script can be invoked
periodically or just once per start of the Splunk host process. In order for
configuration changes to take effect you should restart the Splunk host
process.

When Splunk process needs the modular input script to terminate it will send
the following signal to the process: `SIGTERM` on Unix-like systems and
`CTRL_BREAK` on Windows.

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
