# Splunk Technical Add-on for Splunk OpenTelemetry Collector

The `Splunk_TA_otel` is the replacement for the `Splunk_TA_OTel`.

## Installation

It follows the standard installation process for Splunk apps and add-ons.
Notice that the technical add-on is not a standalone application, it is meant
to be run under `splunkd` control.

## Configuration

See the [inputs.conf.spec](./assets/README/inputs.conf.spec) file for the available
configuration options.

### Splunk Modular Input Primer

This section is a very brief introduction to the general architecture of a
Splunk Modular Input. In the Splunk documentation there is some overlap with
a Splunk App and Splunk Add-on.

A Splunk Modular Input is shipped as a `.tgz` file with a specific [directory
structure](https://docs.splunk.com/Documentation/Splunk/latest/AdvancedDev/ModInputsBasicExample)
in which the root directory is the name of the modular input.

The directory structure will contain [platform and architecture](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/custominputs/modinputsoverview/#Platform-and-architecture-support)
specific paths for the modular _script_, which can be a binary executable,
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
document to the modular input script's `stdin`. This XML is generated from
the `<modular_input>/default/inputs.conf` file and with overrides at the
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
