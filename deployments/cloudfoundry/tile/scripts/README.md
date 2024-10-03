# Tanzu Tile Scripts

## setup_tanzu.sh

This script is used to setup your Tanzu environment for testing with the Splunk OpenTelemetry Collector Tanzu Tile. Running
this script will allow you to install the Tanzu Tile in your Tanzu environment.

**Sample usage:**
```shell
# The hammer file is available in the Self Service center for your specific
# Tanzu environment.
$ export TAS_JSON=/path/to/hammer/file
$ ./setup_tanzu.sh
```

## generate_osdf.py

This script is used to generate the open source disclosure file (OSDF). This is a file that discloses all of the Tanzu
Tile's dependencies. The Tanzu team used to provide a website interface that would generate this file in a specific
format but dropped support for it. This script generates the file in the same format as their website did.

**Sample usage:**
```shell
$ python3 generate_osdf.py --otelcol_version 0.X.0
```

## install_cli_dependencies.sh

This is a shell script used to install all of the CLI tools that are arequired for development
working on the Tanzu Tile. This script was made for MacOS, and has not been
tested on any other OS at this time.