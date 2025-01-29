# Development Guide

## Getting Started

### Start the BOSH Director

[Local machine prerequisites](https://bosh.io/docs/quick-start/#prerequisites)
- The latest VirtualBox environment v7 is incompatible with this
  functionality. Downgrade VirtualBox to [v6](https://www.virtualbox.org/wiki/Download_Old_Builds_6_1) to
  ensure it can work. Tested successfully on `v6.1.42`


Automated start process:
```shell
# Sets up BOSH Director locally
make
# Makes sure local shell gets proper credentials to access director
source bosh-env/virtualbox/.envrc
```

Manual start process:
[Follow the quick start guide to run a BOSH Director.](https://bosh.io/docs/quick-start/)

Common Errors:
- "Waiting for the agent on VM" timeouts

  - Run ```make reinstall-director```

Delete director:
```shell
make delete-director
```
### Upload Ubuntu/OS blob

Note: If you ran ```make``` successfully, you can skip this step.

- [Uploading stemcells guide](https://bosh.io/docs/uploading-stemcells/)
- [Official BOSH stemcells references (including SHAs)](https://bosh.io/stemcells).
- The following is an example command to upload the Warden (BOSH Lite) Ubuntu Bionic (18.04.6 LTS) stemcell:

```shell
bosh upload-stemcell --sha1 d44dc2d1b3f8415b41160ad4f82bc9d30b8dfdce \
https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-bionic-go_agent?v=1.71
````

### Supplemental Documentation Links

[BOSH Release Documentation (Overview)](https://bosh.io/docs/create-release/)

[Quick Start Guide](https://bosh.io/docs/bosh-lite/)

## Create Local BOSH Release

```shell
# Check release script for more environment variables that can be set.
export IS_DEV_RELEASE=1
./release
bosh -e <director-name> upload-release latest-release.tgz
```

## BOSH Release Usage

```shell
# Deploy BOSH Release
bosh -e <director-name> -d splunk-otel-collector deploy deployment.yaml
# Delete BOSH Release
bosh -e <director-name> delete-deployment -d splunk-otel-collector
```
Further explanation of the `deployment.yaml` file is found [here.](#deployment-config)

## Properties

Refer to [./jobs/splunk-otel-collector/spec](jobs/splunk-otel-collector/spec) for the full list of properties
and their descriptions. Properties are provided to the release by using a `deployment.yaml` file.

The following requirements are only true if no OpenTelemetry configuration file is provided
by the user. The user may provide a configuration file by setting the `otel.config_yaml` property.
An example of using this method of providing configuration can be found
[here](./example/custom_config_deployment.yaml).
If no configuration file is provided, a template file will be populated using the following properties.

Required Properties:

- `cloudfoundry.deployment.hostname`
- `cloudfoundry.rlp_gateway.endpoint`
- `cloudfoundry.uaa.endpoint`
- `cloudfoundry.uaa.password`
- `cloudfoundry.uaa.username`
- `splunk.access_token`

- The Splunk Observability Suite requires an endpoint, so one of the following two options must be
  specified:
1) `splunk.api_url` and `splunk.ingest_url`
2) `splunk.realm`

Optional Properties:

- `cloudfoundry.rlp_gateway.shard_id`
    - Default: `opentelemetry`
- `cloudfoundry.rlp_gateway.tls.insecure_skip_verify`
    - Default: `false`
- `cloudfoundry.uaa.tls.insecure_skip_verify`
    - Default: `false`

## Deployment Config

The documentation for a deployment config can be found [here.](https://bosh.io/docs/manifest-v2/)
The deployment config is used when deploying a release, and can provide property values to the deployment.
Example files have been included [here](./example/deployment.yaml) and
[here.](./example/custom_config_deployment.yaml)

### Defining Property Names
Property names are formed by using indentation, not by directly using periods.
To specify `splunk.access_token` in a deployment file:

Correct format:
```yaml
splunk:
  access_token: "..."
```
Incorrect:
```yaml
splunk.access_token: "..."
```

## Debugging

### Useful BOSH CLI Commands
Note: Depending on configuration, all `bosh` commands may require the director name to be explicitly provided.
This is the `-e <director-name>` option.
```shell
# Ensure director is up and running. Director name is "vbox" if created using automated script.
$ bosh -e <director-name> env

# View all bosh releases
$ bosh releases

# View all bosh deployments
$ bosh deployments

# View all bosh VMs. This will show if a deployment's VM is running or failed.
$ bosh vms

# View debug logs for a task (e.g. upload or deploy) that failed
$ bosh task task_number --debug

# View logs for a given deployment. Downloads a TAR file from the deployment.
$ bosh logs -d <deployment-name>

# SSH into a deployment's VM.
$ bosh ssh -d <deployment-name>
```
If VM exists (even in failing state) but SSH fails, it's likely a routing error.
Check the [Quick Start Guide](https://bosh.io/docs/bosh-lite/) document for the directions
on how to set up a local route for the `bosh ssh` command.

### Useful VM Debugging Commands

```shell
# After SSH'ing into VM, see if collector process is running on VM
$ ps aux
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
...
root       19531  2.4  1.7 822180 105656 ?       S<l  08:36   0:55 /var/vcap/packages/splunk_otel_collector/splunk_otel_collector --config /var/vcap/jobs/splunk-otel-collector/bin/config/otel-collector-config.yaml
...

# If collector isn't running, attempt to manually start it to see what (if any) error
# occurred.
$ exec /var/vcap/packages/splunk_otel_collector/splunk_otel_collector \
>   --config /var/vcap/jobs/splunk-otel-collector/bin/config/otel-collector-config.yaml

# Useful VM log files:
/var/vcap/sys/log/splunk-otel-collector/splunk-otel-collector.stdout.log
/var/vcap/sys/log/splunk-otel-collector/splunk-otel-collector.stderr.log

# Proxy settings
/var/vcap/jobs/splunk-otel-collector/bin/config/splunk-otel-collector.conf
```
