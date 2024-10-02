# Tanzu Tile Scripts

## setup_tanzu.sh

This script is used to setup your Tanzu environment for testing with the Splunk OpenTelemetry Collector Tanzu Tile. Running
this script will allow you to install the Tanzu Tile in your Tanzu environment.

**Sample usage:**
```shell
$ export TAS_JSON=/path/to/hammer/file
$ ./setup_tanzu.sh
```

## generate_osdf.py

This script is used to generate the open source disclosure file (OSDF). This is a file that discloses all of the Tanzu
Tile's dependencies. The Tanzu team used to provide a website interface that would generate this file in a specific
format but dropped support for it. This script generates the file in the same format as their website did.

**Note:** This script must be run from the top-level directory of this repository to properly account for the
`license_finder` tool's contents.

**Sample usage:**
```shell
# When creating the tile locally the version can be found in deployments/cloudfoundry/tile/tile-history.yml
$ python3 deployments/cloudfoundry/tile/scripts/generate_osdf.py --otelcol_version 0.X.0
```