# Splunk OpenTelemetry Collector Tanzu Tile

This readme covers the following steps:

- Building 
- Testing
- Releasing

of a Tanzu tile of the [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector).
The Tanzu tile uses the BOSH release to deploy the collector as a [loggregator firehose nozzle](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/nozzle.html).

A crucial step here is the testing and for that you will need to have access to a Tanzu environment and there are a few async tasks that can take some time because it needs approval from Broadcom/VMWare. Follow the steps outlined in the Tanzu ISV Partner Guide to register for a Broadcom Portal Account requesting access to the ISV dashboard. 

Once you have access to the ISV dashboard 

## Create a new TAS environment
1. Get access to Pivotal Partners Slack
2. Create a new TAS environment via: https://self-service.isv.ci/

# Build and Test a Tanzu Tile

Before building and testing a tile, make sure you've updated your repo. Also note that these instructions are specific to Mac OS. The steps are consistent but the binaries will have to be for your OS

## Install Dependencies

1. Download the [Tile Generator](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/tile-generator.html)
  - Note that MacOS support was dropped for this tool, so an older version must be downloaded for darwin development. Version `14.0.6-dev.1` has been confirmed to be working.
  - Download the tile and pcf release assets from [Tile Generator v14.0.6](https://github.com/cf-platform-eng/tile-generator/releases/tag/v14.0.6-dev.1)
    -  you are looking for `tile_darwin-64bit` and `pcf_darwin-64bit`

2. Install [Bosh CLI](https://bosh.io/docs/cli-v2-install/) 
3. Run the install cli script to install other dependencies [script](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/cloudfoundry/tile/scripts/install_cli_depencies.sh). This will make the tile and pcf tools executables and put them in the right directory
  - Note: The downloaded binaries from step 1 should be in the `~/Downloads` folder for this script to work correctly

## Build the Tanzu Tile 

- Go to the tile directory

```
cd splunk-otel-collector/deployments/cloudfoundry/tile
```
- Run the make-latest-tile script
```
$ ./make-latest-tile
```
This creates a file with a `.pivotal` extension in the product directory, which is the tile packaged as a compressed file. Additionally, it creates 2 files named `tile-history.yml` and `tile.yml`. The `tile-history.yml` file has the version of the the tile you will be releasing.

## Configuring the Tile in the Tanzu Environment

The Tanzu Tile created must be imported, configured, and deployed in your Tanzu environment for testing. The import and configuration process can be done via the Tanzu Ops Manager UI, as described below. 

The UI looks like [this](https://github.com/signalfx/splunk-otel-collector/blob/e88b6adb3eafa6076dc0ba94ca1fa742b5830bf5/deployments/cloudfoundry/tile/resources/tanzu_tile_in_ops_mgr.png) 
The Configuration Page looks like [this](https://github.com/signalfx/splunk-otel-collector/blob/e88b6adb3eafa6076dc0ba94ca1fa742b5830bf5/deployments/cloudfoundry/tile/resources/tanzu_tile_config_options.png)

### Ops Manager Configuration

- Browse to the Tanzu Ops Manager that you've created in the self service environment.
- Login using credentials provided in self service center.
- Upload Tanzu Tile
  - Click`IMPORT A PRODUCT` -> select the created Tanzu Tile.
  - The Tile will show up in the left window pane, simply click `+` next to it.
  - The Tile will be shown on the Installation Dashboard as not being configured properly.
- Configure Tanzu Tile
  - Assign AZs and Networks - Fill in required values, these do not have an impact on the Collector's deployment.
  - Nozzle Config - The two UAA arguments are required, use the values supplied by the [setup script]([./scripts/setup_tanzu.sh](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/tile/scripts#setup_tanzush)). The default username is `my-v2-nozzle` and password is `password`.
  - Splunk Observability Cloud - These config options are directly mapped to the SignalFx's exporter options, so fill in values you use there.
  - Resource Config - No changes necessary.
  - Click Save after every page's changes.


## Test the Tanzu Tile

- Since you've already created a TAS environment, download the hammer file from the ISV UI. its a JSON file 
- Configure your Tanzu environment by running this [script](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/cloudfoundry/tile/scripts/setup_tanzu.sh)
- Deploying the Tanzu Tile
  - Install the Tanzu Tile
  - Browse to Tanzu Ops Manager home page.
  - Ensure Tanzu Tile shows up with a green bar underneath (configuration is complete).
  - Select REVIEW PENDING CHANGES
    - Optional: Unselect Small Footprint VMware Tanzu Application Service. Unselecting this will speed up deployment time.
  - Select APPLY CHANGES
    - Note that this takes a few minutes and a progress will show the status
- Unless you run into issues listed in the [common issues section](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/cloudfoundry/tile/DEVELOPMENT.md#ops-manager-configuration) data should now be flowing into Splunk
- Go the realm you are sending data to and open the OOTB dashboard by searching for 'VMware' or 'Tanzu'
  - The `VMware Tanzu AS` dashboard should have a chart called containers that should show your Tanzu deployment name
  - If you see said chart, congratulations!  Your tile works, and is ready to be released into the wild via the release process outlined next 
    
## Releasing

The next few steps will be run from the root directory of the Splunk OpenTelemetry Collector

- You've already installed the `license_finder` tool at this point. Add the known licenses using the tool. Run the `add_known_licenses.sh` scripts located under `/tile/scripts`.
- Create an OSDF (open source disclosure file).
  - Use license_finder to determine if there are any new dependencies by running the following command:

```shell
$ license_finder # Will show if any dependencies need to be approved
Dependencies that need approval:
```

- If you don’t see the "Dependencies that need approval" message
  - Use `generate_osdf.py` script to generate OSDF file. Output will be in the local file `OSDF VX.XX.X.txt`
  ```
    python deployments/cloudfoundry/tile/scripts/generate_osdf.py --otelcol_version <version from tile-history.yml>
  ```
- Otherwise, you will need to manually approve each dependency. This will likely follow the format of the original license_finder commands, but you’ll have to determine which license applies to a given dependency, or if a new license can be approved or not.
  - For guidance on whether a license is approved for use, reach out to your product, legal or the Broadcom team through the Pivotal Partners Slack
  - Any commands that needed to be run should be added to the list of commands above, so future users don’t have to duplicate your work.

- Next step is to make the actual release by uploading the artifiacts and filling out a form. The process is outlined in the Tanzu ISV Partner Guide
