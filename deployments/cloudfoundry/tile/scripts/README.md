# Splunk OpenTelemetry Collector Tanzu Tile

This readme covers the following steps 

- Building 
- Testing
- Releasing

of a Tanzu tile of the [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector).
The Tanzu tile uses the BOSH release to deploy the collector as a [loggregator firehose nozzle](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/nozzle.html).

A crucial step here is the testing and for that you will need to have access to a Tanzu environment and there are a few async tasks that can take some time because it needs approval from Broadcom/VMWare. So make sure to complete the steps listed here [TAS setup](https://github.com/signalfx/signalfx-agent/tree/main/pkg/monitors/cloudfoundry#create-a-new-tas-environment) to create a new TAS environment. You will need this very soon

# Build and Test

Before building and testing a tile, make sure you've updated your repo. Also note that these instructions are specific to Mac OS. The steps are consistent but the binaries will have to be  for your OS

## Install Dependencies

1. Download the tile generator [Tile Generator](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/tile-generator.html)
  - Note that MacOS support was dropped for this tool, so an older version must be downloaded for darwin development. Version `14.0.6-dev.1` has been confirmed to be working.
  - Download the tile and pcf release assets from [Tile Generator v14.0.6](https://github.com/cf-platform-eng/tile-generator/releases/tag/v14.0.6-dev.1)
    -  you are looking for tile_darwin-64bit and pcf_darwin-64bit

2. Install [Bosh CLI](https://bosh.io/docs/cli-v2-install/) 
3. Run the install cli script to install other dependencies [script](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/cloudfoundry/tile/scripts/install_cli_depencies.sh). This will make the tile and pcf tools executables and put them in the right directory

## Build the Tanzu Tile 

- Go to the tile directory

```
cd splunk-otel-collector/edit/main/deployments/cloudfoundry/tile
```
- Run the make-latest-tile script
```
$ ./make-latest-tile
```
This creates a file with a .pivotal extension in the product directory. Additionally, it creates 2 files called tile-history.yml and tile.yml. The tile-history.yml file has the version of the the tile you will be releasing.

## Test the Tanzu Tile

- Since you've already created a TAS environment, download the hammer file from the ISV UI. its a JSON file 
- Configure your Tanzu environment by running this [script](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/cloudfoundry/tile/scripts/setup_tanzu.sh)
- Unless you run into issues listed in the [common issues section](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/cloudfoundry/tile/DEVELOPMENT.md#ops-manager-configuration) data should be flowing into Splunk
- Go the realm you are sending data to and open the OOTB dashboard by searching for 'VMware' or 'Tanzu'
  - The Tanzu AS overview dashboard should have a chart called containers that should show your Tanzu deployment name
  - Congrats your tile works. Its ready to be released into the wild 
    
## Releasing

The next few steps will be run from the root directory of the Splunk OpenTelemetry Collector

- You've already installed the license_finder tool at this point. Add the known licenses using the tool. Copy paste the contents of add_known_licenses.txt to your terminal and hit enter
- Create an OSDF (open source disclosure file).
  - Use license_finder to determine if there are any new dependencies by running the following command:

```shell
$ license_finder # Will show if any dependencies need to be approved
Dependencies that need approval:
```

- If you don’t see the Dependencies that need approval message
  - Use generate_osdf.py script to generate OSDF file. Output will be in the local file “OSDF VX.XX.X.txt”
  ```shell
    python deployments/cloudfoundry/tile/scripts/generate_osdf.py --otelcol_version <version from tile-history.yml>
  ```
- Otherwise, you will need to manually approve each dependency. This will likely follow the format of the original license_finder commands, but you’ll have to determine which license applies to a given dependency, or if a new license can be approved or not.
  - For guidance on whether a license is approved for use, reach out to Aunsh, legal, or the Broadcom team through the Pivotal Partners Slack using the private #signalfx channel.
  - Any commands that needed to be run should be added to the list of commands above, so future users don’t have to duplicate your work.

- Next step is to make the actual release by uploading the artifiacts and filling out a form. The process is outlined [here](https://drive.google.com/file/d/1lIJly4qS4drsE0jhmk6ZcgA3AS80hzWP/view) under `Publishing new releases of your product`
  - You need the .pivotal file and the osdf file at a minimum
  - For Tanzu Tile software dependencies:
    - Stemcell: Go to Tanzu Ops Manager, the environment spun up by the self service center. Browse to STEMCELL LIBRARY. If you’ve uploaded your Tanzu Tile you should see which stemcell your Tanzu Tile is dependent upon, and you can use this information to specify to users which stemcell is required.
    - Ops manager: Go to the self service center. Search in the Credentials: section for ops_manager_version. This can be used to specify which Ops manager version our tile is compatible with.

